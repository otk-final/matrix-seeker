package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"log"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"matrix-seeker/seeker"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func InitCli() *cli.App {

	app := cli.NewApp()
	app.Name = "矩阵爬虫 - 终端"

	app.Version = "1.0.0"
	app.UsageText = "加载本地指定脚本文件执行"

	app.Commands = []cli.Command{{
		Name:   "start",
		Usage:  "执行脚本",
		Action: startCmd,
	}, {
		Name:   "reload",
		Usage:  "重新加载执行",
		Action: reloadCmd,
	}}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return app
}

func initContext(scriptPath string) (*seeker.FetchContext, *artifact.Persistent, *os.File) {
	//解析root脚本
	cfg := &meta.FetchConfig{
		ScriptPath: scriptPath,
		TimeOut:    time.Second * 60,
	}

	//初始化上下文
	ft := &seeker.FetchContext{
		Config: cfg,
		//构建信道
		WideChan:   make(chan *seeker.WideHandler, 1),
		DepthChan:  make(chan *seeker.DepthHandler, 1),
		MatrixChan: make(chan *seeker.MatrixHandler, 1),
		//全局锁，标识所有连接都执行完成
		CtxWait: &sync.WaitGroup{},
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

	//判断文件夹是否存在
	if _, err := os.Stat(at.OutputDir); os.IsNotExist(err) {
		os.MkdirAll(at.OutputDir, os.ModePerm)
	}

	logFile, err := os.OpenFile(at.OutputDir+"/"+"seeker.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("日志文件加载失败:[%s]", err.Error()))
		return nil, nil, nil
	}

	//设置目录
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	return ft, at, logFile
}

//执行
func startCmd(c *cli.Context) {

	//参数
	scriptPath := c.Args().First()

	//初始化
	ft, at, logFile := initContext(scriptPath)
	defer logFile.Close()

	//创建根节点
	root := script.CreateLinkNode(scriptPath, "root.json")

	//执行
	ft.Start(root, at)
}

//重新加载执行
func reloadCmd(c *cli.Context) {

	jsonPath := c.Args().First()
	scriptPath := c.Args().Get(1)

	//判断文件是否存在
	jsonFile, err := os.Stat(jsonPath)
	if os.IsNotExist(err) {
		return
	}

	//加载数据
	nodeName, defaultPath, allData := loadFetchData(jsonPath, jsonFile)

	//获取脚本路径 如果未配置脚本路径，取默认的
	if scriptPath == "" {
		scriptPath = defaultPath
	}
	if scriptPath == "" {
		return
	}

	//通过文件目录名称，获取当前节点
	linkNode := script.FindLinkNode(scriptPath, nodeName)
	if linkNode == nil {
		return
	}

	//初始化
	ft, at, logFile := initContext(scriptPath)
	defer logFile.Close()

	//执行
	ft.Reload(linkNode, allData, at)
}

func downloadCmd(c *cli.Context) {

}

func loadFetchData(jsonPath string, jsonFile os.FileInfo) (string, string, []*meta.FileFetchData) {

	loadData := func(path string) *meta.FileFetchData {
		//判断文件夹是否存在
		file, err := os.Open(path)
		if err != nil && os.IsNotExist(err) {
			panic(err)
		}
		defer file.Close()

		//读取文件
		dc := json.NewDecoder(file)

		//序列化
		var fileData meta.FileFetchData
		err = dc.Decode(&fileData)
		if err != nil {
			log.Println("加载异常")
		}
		return &fileData
	}

	//返回数据
	allData := make([]*meta.FileFetchData, 0)
	nodeName := ""
	scriptPath := ""
	if jsonFile.IsDir() {
		//目录 执行目录下所有文件
		nodeName = jsonFile.Name()
		//遍历
		ra, _ := ioutil.ReadDir(jsonPath)
		for _, f := range ra {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			allData = append(allData, loadData(jsonPath+"/"+f.Name()))
		}
		scriptPath = filepath.Join(jsonPath, "../..")
	} else {

		nodeFile, _ := os.Stat(filepath.Dir(jsonPath))
		nodeName = nodeFile.Name()

		//文件 执行单个文件
		allData = append(allData, loadData(jsonPath))
		scriptPath = filepath.Join(jsonPath, "../../..")
	}

	return nodeName, scriptPath, allData
}
