package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"matrix-seeker/seeker"
	"runtime"
	"time"
)

func main1() {

	termbox.Init()
	defer termbox.Close()

Loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				fmt.Println("You press Esc")
			case termbox.KeyF1:
				fmt.Println("You press F1")
			default:
				break Loop
			}
		}
	}

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

	//存储(本地存储)
	at := &artifact.Persistent{
		WaitNode: make(chan *meta.FetchNode, 1),
	}

	//执行
	ft.Execute(root, at)
}
