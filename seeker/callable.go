package seeker

import (
	"context"
	"fmt"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func (f *FetchContext) Execute(root *meta.FetchNode, artifact *artifact.Artifact) {

	wg := sync.WaitGroup{}

	//构建信道
	f.wideChan = make(chan *WideHandler, 0)
	f.depthChan = make(chan *DepthHandler, 0)
	f.matrixChan = make(chan *MatrixHandler, 0)

	//开启执行任务
	wg.Add(2)

	//初始请求
	go func(node *meta.FetchNode) {
		defer wg.Done()
		//启动根节点（深度）
		rootCtx, _ := context.WithCancel(context.Background())

		//优先读取脚本初始化请求
		req := script.CreateRequest(node, nil, f.JsVm, "infoRoot", nil)
		if req == nil {
			url, err := url.ParseRequestURI(f.Config.HttpUrl)
			if err != nil {
				return
			}
			//构建请求
			req = &http.Request{
				Method: "GET",
				URL:    url,
			}
		}

		//执行
		f.depthChan <- f.CreateDepthHandler(rootCtx, node, req)
	}(root)

	//开启监控任务
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(f.Config.TimeOut):
				fmt.Println("wait...........")
			case mc := <-f.matrixChan: //矩阵
				go mc.Fetch()
			case dc := <-f.depthChan: //深度
				go dc.Fetch()
			case wc := <-f.wideChan: //广度
				go wc.Fetch()
			default:
				break
			}
		}
	}()

	wg.Wait()
	//开启值处理任务

}
