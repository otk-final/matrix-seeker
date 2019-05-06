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
	}, {
		Name:   "download",
		Usage:  "下载素材图片",
		Action: downloadCmd,
	}}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return app
}

//执行
func startCmd(c *cli.Context) {

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

	//判断文件夹是否存在
	if _, err := os.Stat(at.OutputDir); os.IsNotExist(err) {
		os.MkdirAll(at.OutputDir, os.ModePerm)
	}

	logFile, err := os.OpenFile(at.OutputDir+"/"+"seeker.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("日志文件加载失败:[%s]", err.Error()))
		return
	}
	defer logFile.Close()

	//设置目录
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	//执行
	ft.Execute(root, at)

	ft = nil
	root = nil
	at = nil
}

//重新加载执行
func reloadCmd(c *cli.Context) {
	jsonPath := c.Args().First()

	//判断文件是否存在
	jsonFile, err := os.Stat(jsonPath)
	if os.IsNotExist(err) {
		return
	}
	//加载数据
	nodeName, allData := loadFetchData(jsonPath, jsonFile)
	//通过文件目录名称，获取当前节点
	linkNode := script.FindLinkNode(jsonPath, nodeName)

	log.Println(allData)
	log.Println(linkNode)
}

func downloadCmd(c *cli.Context) {

}

func loadFetchData(jsonPath string, jsonFile os.FileInfo) (string, []*meta.FileFetchData) {

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
	} else {
		nodeName = filepath.Dir(jsonPath)
		//文件 执行单个文件
		allData = append(allData, loadData(jsonPath))
	}

	return nodeName, allData
}
