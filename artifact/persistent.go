package artifact

import (
	"container/list"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"log"
	"matrix-seeker/meta"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Persistent struct {
	OutputDir string
	WaitGroup *sync.WaitGroup
}

func (at *Persistent) bulkOf(node *meta.FetchNode) {
	defer at.WaitGroup.Done()
	/*
		优先判断数据
	*/
	if len(node.Data) == 0 {
		return
	}

	//根据当前节点名称，定位目录结构
	paths := []string{at.OutputDir, node.Name, uuid.NewV4().String() + ".json"}
	nodePath := strings.Join(paths, "/")

	//fmt.Println(fmt.Sprintf("Referer[%s]", node.Referer))

	//获取当前节点目录（创建imgOut目录用户下载图片）
	downloadArray := node.GetActionValues("download")
	//待所有图片图片下载完成
	currDir := filepath.Dir(nodePath)
	for _, dv := range downloadArray {
		at.WaitGroup.Add(1)
		//遍历出当前结果集中需要下载图片的字段，名称
		go at.CreateImgTask(currDir+"/素材/", node.Referer, dv)
	}

	//将内容转换为json格式存储
	nodeByte, err := json.Marshal(node.Data)
	if err != nil {
		return
	}

	log.Println(fmt.Sprintf("生成文件[%s]", nodePath))
	//写入文件
	waitToFile(nodePath, nodeByte)
}

func waitToFile(filePath string, data []byte) {

	fileDir := filepath.Dir(filePath)

	//判断文件夹是否存在
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		os.MkdirAll(fileDir, os.ModePerm)
	}

	//将node.Data输出至文件(覆盖模式）
	nodeFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer nodeFile.Close()

	nodeFile.Write(data)
}

//采用深度优先（子节点放入栈首）
func (at *Persistent) Bulk(node *meta.FetchNode) {

	//模拟栈
	stack := list.New()

	//根节点入栈
	stack.PushFront(node)

	for {

		if stack.Len() == 0 {
			break
		}

		//出栈(删除)
		tmp := stack.Front()
		tmpVal := stack.Remove(tmp).(*meta.FetchNode)

		//执行单个节点任务
		at.WaitGroup.Add(1)
		go at.bulkOf(tmpVal)

		//下一个节点
		if len(tmpVal.Childrens) > 0 {
			for _, child := range tmpVal.Childrens {
				stack.PushFront(child)
			}
		}
	}

	at.WaitGroup.Wait()
}
