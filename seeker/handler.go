package seeker

import (
	"fmt"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"net/http"
	"net/url"
	"sync"
	"time"
)

/*

 */

func (f *FetchContext) Execute(root *meta.FetchNode, ap *artifact.Persistent) {

	if root.Name == "" {
		root.Name = "ROOT"
	}

	//构建信道
	f.wideChan = make(chan *WideHandler, 1)
	f.depthChan = make(chan *DepthHandler, 1)
	f.matrixChan = make(chan *MatrixHandler, 1)
	//全局锁，标识所有连接都执行完成
	f.ctxWait = &sync.WaitGroup{}

	//初始请求(同步)
	f.startRoot(root)

	//开启监控任务（异步消费）
	go f.monitor()

	//任务是否完成
	f.finish()

	//通知持久化
	ap.Bulk(root)
}

func (f *FetchContext) startRoot(root *meta.FetchNode) {

	//优先读取脚本初始化请求
	req := script.CreateRequest(root, nil, f.JsVm, "startRoot", nil)
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
	f.depthChan <- f.CreateDepthHandler(root, req)
}

func (f *FetchContext) finish() {

	//等待所有任务行为完成
	f.ctxWait.Wait()

	close(f.wideChan)
	close(f.depthChan)
	close(f.matrixChan)

	f.finished = true

}

func (f *FetchContext) monitor() {

	for {

		//关闭状态（关闭后进行退出）
		if f.finished {
			break
		}

		select {
		case <-time.After(time.Second * 10):
			fmt.Println("-------------------------wait-------------------------")
		case mc, ok := <-f.matrixChan: //矩阵
			if ok {
				go mc.Fetch()
			}
		case dc, ok := <-f.depthChan: //深度
			if ok {
				go dc.Fetch()
			}
		case wc, ok := <-f.wideChan: //广度
			if ok {
				go wc.Fetch()
			}
		default:
			break
		}
	}

}
