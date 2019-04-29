package main

import (
	"bufio"
	"github.com/urfave/cli"
	"log"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"matrix-seeker/seeker"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func initCli() *cli.App {

	app := cli.NewApp()
	app.Name = "矩阵爬虫 - 终端"

	app.Version = "1.0.0"
	app.UsageText = "加载本地指定脚本文件执行"

	app.Commands = []cli.Command{{
		Name:   "start",
		Usage:  "执行脚本",
		Action: startCmd,
	}}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return app
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	app := initCli()
	//监控用户输入
	for {
		var input string

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		//构建命令
		s := []string{app.Name}

		//获取命令
		cmdArgs := strings.Split(input, " ")
		if len(cmdArgs) == 0 {
			continue
		}

		s = append(s, cmdArgs...)
		app.Run(s)
	}
}

//执行
func startCmd(c *cli.Context) error {
	scriptPath := c.Args().First()

	//解析root脚本
	cfg := &meta.FetchConfig{
		ScriptPath: scriptPath,
		TimeOut:    time.Second * 60,
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
		OutputDir: scriptPath + "/out",
		WaitGroup: &sync.WaitGroup{},
	}

	//执行
	ft.Execute(root, at)

	return nil
}
