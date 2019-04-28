package script

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"log"
	"matrix-seeker/meta"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var lock = &sync.Mutex{}

func CreateNodeName(node *meta.FetchNode, fromReq *http.Request, scriptVm *otto.Otto, funcName string, args interface{}) string {

	//必须采用锁机制（otto 的 bug)
	lock.Lock()
	defer lock.Unlock()

	//执行js函数
	val, err := exeFunc(node, fromReq, scriptVm, funcName, args)
	if err != nil {
		return ""
	}

	return val.String()
}

func CreateRequest(node *meta.FetchNode, fromReq *http.Request, scriptVm *otto.Otto, funcName string, args interface{}) *http.Request {

	//必须采用锁机制（otto 的 bug)
	lock.Lock()
	defer lock.Unlock()

	//执行js函数
	val, err := exeFunc(node, fromReq, scriptVm, funcName, args)
	if err != nil {
		return nil
	}

	return cvtRequest(val)
}

/*
	创建分页请求
*/
func exeFunc(node *meta.FetchNode, fromReq *http.Request, scriptVm *otto.Otto, funcName string, args interface{}) (otto.Value, error) {

	//检查方法是否存在，并有效
	_, err := scriptVm.Get(funcName)
	if err != nil {
		return otto.NullValue(), nil
	}

	//对参数做序列化
	nodeByte, _ := json.Marshal(node)
	nodeJson, _ := scriptVm.Eval("(" + string(nodeByte) + ")")

	//当前请求
	reqJson := cvtValue(scriptVm, fromReq)

	//判断值类型(两种）
	argValue, _ := cvtArgs(scriptVm, args)

	//将当前请求来源传递给用户自定义函数
	return scriptVm.Call(funcName, nil, nodeJson, reqJson, argValue)

}

func CreateLinkNode(scriptDir string, fileName string) *meta.FetchNode {

	//判断文件夹是否存在
	file, err := os.Open(scriptDir + "/" + fileName)
	if err != nil && os.IsNotExist(err) {
		panic(err)
	}
	defer file.Close()

	//读取文件
	dc := json.NewDecoder(file)

	//序列化
	var jsonNode meta.FetchNode
	err = dc.Decode(&jsonNode)
	if err != nil {
		panic(err)
	}

	return &jsonNode
}

/*
	加载脚本文件
*/
func LoadContext(fileDir string, scriptName string) (*otto.Otto, error) {

	//判断文件夹是否存在
	file, err := os.Open(fileDir + "/" + scriptName)
	if err != nil && os.IsNotExist(err) {
		log.Println(fmt.Sprintf("脚本文件[%s]不存在:", scriptName))
		return nil, err
	}
	defer file.Close()

	//读取文件
	by, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(fmt.Sprintf("脚本文件[%s]读取异常:", scriptName))
		return nil, err
	}

	//脚本
	script := string(by)
	if script == "" {
		log.Println(fmt.Sprintf("脚本文件[%s]内容为空:", scriptName))
		return nil, errors.New("脚本文件为空")
	}

	//javascript 虚拟执行环境
	vm := otto.New()
	ok, err := vm.Run(script)

	fmt.Println(ok)
	if err != nil {
		log.Println(fmt.Sprintf("脚本文件[%s]初始化异常:[%s]", scriptName, err.Error()))
		return nil, err
	}

	return vm, nil
}

func cvtArgs(scriptVm *otto.Otto, argVal interface{}) (otto.Value, error) {

	if argVal == nil {
		return otto.NaNValue(), nil
	}

	switch argVal.(type) {
	case int:
		return otto.ToValue(argVal)
	case []*meta.FetchData:
		array := argVal.([]*meta.FetchData)
		fmtMap := make(map[string]interface{}, 0)
		for _, d := range array {
			fmtMap[d.Field] = d.Value
		}
		//对参数做序列化
		argByte, _ := json.Marshal(fmtMap)
		return scriptVm.Eval("(" + string(argByte) + ")")
	default:
		return otto.UndefinedValue(), nil
	}
}

/*
	将golang的request对象转换为js对象
	{
		url:"",
		method:"",
		params:{},
		header:{}
	}
*/

func cvtValue(scriptVm *otto.Otto, r *http.Request) otto.Value {
	if r == nil {
		return otto.UndefinedValue()
	}

	valMap := map[string]interface{}{
		"url":    r.URL.String(),
		"header": r.Header,
		"method": r.Method,
		"params": r.Form,
	}
	//json序列化
	valByte, _ := json.Marshal(valMap)
	//强转js 的json对象
	valJson, _ := scriptVm.Eval("(" + string(valByte) + ")")
	return valJson
}

func cvtRequest(val otto.Value) *http.Request {
	if val.IsUndefined() || !val.IsObject() {
		return nil
	}

	req := &http.Request{}
	instance := val.Object()

	//URL
	urlVal, err := instance.Get("url")
	if err == nil {
		url, _ := url.ParseRequestURI(urlVal.String())
		req.URL = url
	}

	//Method
	methodVal, err := instance.Get("method")
	if err == nil && methodVal.IsString() {
		req.Method = methodVal.String()
	} else {
		req.Method = "GET"
	}

	toHeader := func(obj *otto.Object) http.Header {
		keys := obj.Keys()
		out := http.Header{}
		for _, key := range keys {
			_val, err := obj.Get(key)
			if err != nil {
				continue
			}
			out.Add(key, _val.String())
		}
		return out
	}

	toParams := func(obj *otto.Object) url.Values {
		keys := obj.Keys()
		urlVal := url.Values{}

		for _, key := range keys {
			_val, err := obj.Get(key)
			if err != nil {
				continue
			}
			urlVal.Add(key, _val.String())
		}
		return urlVal
	}

	//Params
	paramsVal, err := instance.Get("params")
	if err == nil && paramsVal.IsObject() {
		//转换
		urlValues := toParams(paramsVal.Object())
		//替换
		if strings.EqualFold(req.Method, "GET") {
			req.URL.RawQuery = urlValues.Encode()
		} else {
			req.PostForm = urlValues
		}
	}

	//Header
	headerVal, err := instance.Get("header")
	if err == nil && headerVal.IsObject() {
		req.Header = toHeader(headerVal.Object())
	}

	return req
}
