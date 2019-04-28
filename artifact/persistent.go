package artifact

import (
	"encoding/json"
	"matrix-seeker/meta"
	"os"
	"path/filepath"
	"sync"
)

type Persistent struct {
	OutputDir string
	WaitNode  chan *meta.FetchNode
}

func (at *Persistent) bulkOf(wg *sync.WaitGroup, node *meta.FetchNode, item []*meta.FetchData) {
	defer wg.Done()

	//根据当前节点名称，定位目录结构
	nodePath := node.GetNodeFilePath(at.OutputDir)
	/*
		将node.Data输出至文件(覆盖模式）
	*/
	nodeFile, err := os.OpenFile(nodePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer nodeFile.Close()

	//获取当前节点目录（创建imgOut目录用户下载图片）
	downloadArray := getActions(node, item, "download")
	for _, dv := range downloadArray {
		//遍历出当前结果集中需要下载图片的字段，名称
		at.CreateImgTask(dv, filepath.Dir(nodePath))
	}

	//将内容转换为json格式存储
	nodeByte, err := json.Marshal(node.Data)
	nodeFile.Write(nodeByte)
}

func (at *Persistent) Bulk(node *meta.FetchNode) {
	bw := &sync.WaitGroup{}
	//轮询执行
	for _, item := range node.Data {
		bw.Add(1)
		go at.bulkOf(bw, node, item)
	}
	bw.Wait()
}

func getActions(node *meta.FetchNode, item []*meta.FetchData, actionType string) []string {

	//检查
	fields := make(map[string]interface{}, 0)
	for _, field := range node.Bind.Fields {
		if field.ActionType == actionType {
			fields[field.Mapper] = "ok"
		}
	}

	outs := make([]string, 0)
	for _, data := range item {
		_, ok := fields[data.Field]
		if !ok {
			continue
		}

		//判断类型
		switch data.Value.(type) {
		case string:
			//判断类型
			outs = append(outs, data.Value.(string))
		case []string:
			outs = append(outs, data.Value.([]string)...)
		}
	}

	return outs
}
