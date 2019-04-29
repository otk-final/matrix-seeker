package main

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"matrix-seeker/seeker"
	"os"
	"runtime"
	"sync"
	"time"
)

func main1() {
	pic := "D:/seeker/out/分类/素材/" + "uploads-pic-1-5bac2d8691500_275_275.jpg"

	//检查文件是否存在
	if imgFile, _ := os.Stat(pic); imgFile != nil {
		return
	}

	fmt.Println(uuid.NewV4())
}
func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	//解析root脚本
	cfg := &meta.FetchConfig{
		ScriptPath: "D://DEV/GoProject/matrix-seeker/script-example/七丽时尚",
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

	//存储(本地存储) ,默认当前目录下out文件夹
	at := &artifact.Persistent{
		OutputDir: "D:/seeker/out",
		WaitGroup: &sync.WaitGroup{},
	}

	//执行
	ft.Execute(root, at)
}
