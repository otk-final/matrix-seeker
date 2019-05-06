package seeker

import (
	"errors"
	"fmt"
	"log"
	"matrix-seeker/artifact"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"net/http"
	"net/url"
	"time"
)

func (f *FetchContext) Reload(currNode *meta.FetchNode, loads []*meta.FileFetchData, ap *artifact.Persistent) {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	f.CtxWait.Add(1)
	//遍历执行
	go func() {
		defer f.CtxWait.Done()
		for _, fileData := range loads {

			//反序列化当前req
			url, _ := url.Parse(fileData.Referer)
			req := &http.Request{URL: url}

			f.eachCall(currNode, req, fileData.Data)
		}
	}()

	//开启监控任务（异步消费）
	go f.monitor()

	//任务是否完成
	f.finish()

	//通知持久化
	ap.Bulk(currNode)

	log.Println(fmt.Sprintf("输出[%s]", ap.OutputDir))
	log.Println("结束...")
}

func (f *FetchContext) Start(root *meta.FetchNode, ap *artifact.Persistent) {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	if root.Name == "" {
		root.Name = "ROOT"
	}

	//初始请求(同步)
	f.startRoot(root)

	//开启监控任务（异步消费）
	go f.monitor()

	//任务是否完成
	f.finish()

	//通知持久化
	ap.Bulk(root)

	log.Println(fmt.Sprintf("输出[%s]", ap.OutputDir))
	log.Println("结束...")
}

func (f *FetchContext) startRoot(root *meta.FetchNode) {

	//优先读取脚本初始化请求
	req := script.CreateRequest(root, nil, f.JsVm, "startRoot", nil)
	if req == nil {
		panic(errors.New("脚本初始化请求错误"))
		return
	}

	//执行
	f.DepthChan <- f.CreateDepthHandler(root, req)
}

func (f *FetchContext) finish() {

	//等待所有任务行为完成
	f.CtxWait.Wait()

	close(f.WideChan)
	close(f.DepthChan)
	close(f.MatrixChan)

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
		case mc, ok := <-f.MatrixChan: //矩阵
			if ok {
				go mc.Fetch()
			}
		case dc, ok := <-f.DepthChan: //深度
			if ok {
				go dc.Fetch()
			}
		case wc, ok := <-f.WideChan: //广度
			if ok {
				go wc.Fetch()
			}
		default:
			break
		}
	}

}
