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
)

/*
	创建分页请求
*/
func CreateRequest(node *meta.FetchNode, fromReq *http.Request, scriptVm *otto.Otto, funcName string, args interface{}) *http.Request {

	//检查方法是否存在，并有效
	_, err := scriptVm.Get(funcName)
	if err != nil {
		return nil
	}

	//对参数做序列化
	nodeByte, _ := json.Marshal(node)
	nodeJson, _ := scriptVm.Eval("(" + string(nodeByte) + ")")

	var reqJson otto.Value
	if fromReq != nil {
		reqMap := &map[string]interface{}{
			"Header":   fromReq.Header,
			"URL":      fromReq.URL,
			"Method":   fromReq.Method,
			"Form":     fromReq.Form,
			"PostForm": fromReq.PostForm,
		}
		reqByte, _ := json.Marshal(reqMap)
		reqJson, _ = scriptVm.Eval("(" + string(reqByte) + ")")
	}

	//判断值类型
	argValue, _ := otto.ToValue(args)

	//将当前请求来源传递给用户自定义函数
	value, err := scriptVm.Call(funcName, nil, nodeJson, reqJson, argValue)
	if err != nil {
		return nil
	}

	return fmtCallOut(value)
}

func fmtCallOut(value otto.Value) *http.Request {
	//将value转换为req对象
	out := value.Object()

	req := &http.Request{}

	//URL
	urlVal, _ := out.Get("URL")

	reqUrl := &url.URL{}
	err := json.Unmarshal([]byte(urlVal.String()), *reqUrl)
	if err != nil {
		return nil
	}
	req.URL = reqUrl

	//Method
	methodVal, _ := out.Get("Method")
	if methodVal.String() == "" {
		req.Method = "GET"
	} else {
		req.Method = methodVal.String()
	}

	////Header
	//header := http.Header{}
	//headerVal, err := out.Get("header")
	//
	//
	////Form,PostForm
	//cvtUrlValues := func(name string) url.Values {
	//
	//}
	//
	//req.Form = cvtUrlValues("form")
	//req.PostForm = cvtUrlValues("postForm")
	return req
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
		log.Println(fmt.Sprintf("脚本文件[%s]初始化异常:", scriptName))
		return nil, err
	}

	return vm, nil
}

func fetchDataArrayToMap(array []*meta.FetchData) map[string]interface{} {
	fmtMap := make(map[string]interface{}, 0)
	for _, d := range array {
		fmtMap[d.Field] = d.Value
	}
	return fmtMap
}
