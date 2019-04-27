package main

import (
	"encoding/json"
	"fmt"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"matrix-seeker/seeker"
	"net/url"
	"time"
)

func main1() {

	ur := &url.URL{
		Host: "www.baidu.com",
	}
	reqByte, err := json.Marshal(ur)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(reqByte))

}
func main() {

	//解析root脚本
	cfg := &meta.FetchConfig{
		ScriptPath: "F://seeker/七丽时尚",
		HttpUrl:    "http://www.7y7.com/qinggan/",
		TimeOut:    time.Second * 10,
	}

	root := script.CreateLinkNode(cfg.ScriptPath, "root.json")
	//初始化上下文
	ft := &seeker.FetchContext{
		Config: cfg,
	}

	//加载script文件
	vm, err := script.LoadContext(cfg.ScriptPath, "script.js")
	if err == nil {
		ft.JsVm = vm
	}

	//存储(本地存储)
	at := &artifact.Artifact{}

	//执行
	ft.Execute(root, at)
}
