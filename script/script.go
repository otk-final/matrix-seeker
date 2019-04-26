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
	"os"
)

func fetchDataArrayToMap(array []*meta.FetchData) map[string]interface{} {
	fmtMap := make(map[string]interface{}, 0)
	for _, d := range array {
		fmtMap[d.Field] = d.Value
	}
	return fmtMap
}

/*
	创建分页请求
*/
func CreateRequest(node *meta.FetchNode, cfg *meta.FetchConfig, scriptVm *otto.Otto, funcName string, args interface{}) *http.Request {

	//检查方法是否存在，并有效
	handler, err := scriptVm.Get(funcName)
	if err != nil || !handler.IsFunction() {
		return nil
	}

	value, err := scriptVm.Call(funcName, nil, args)
	if err != nil {
		return nil
	}

	//将value转换为req对象
	value.Object()
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

	/*
		加载公共函数
		公共报文头
	*/
	vm.Set("$setGlobalHeader", nil)

	

	return vm, nil
}
