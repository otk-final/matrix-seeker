package seeker

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
	"golang.org/x/net/html/charset"
	"log"
	"matrix-seeker/meta"
	"matrix-seeker/script"
	"net/http"
	"strings"
	"sync"
	"time"
)

type FetchContext struct {
	Config     *meta.FetchConfig
	JsVm       *otto.Otto
	CtxWait    *sync.WaitGroup
	finished   bool
	WideChan   chan *WideHandler
	MatrixChan chan *MatrixHandler
	DepthChan  chan *DepthHandler
}

/*
	广度
	支持：分页，多页面
*/
type WideHandler struct {
	node  *meta.FetchNode
	Fetch func()
}

/*
	矩阵
	支持：列表
*/
type MatrixHandler struct {
	node  *meta.FetchNode
	Fetch func()
}

/*
	深度
	支持：跳转，脚本获取
*/
type DepthHandler struct {
	node  *meta.FetchNode
	Fetch func()
}

func (f *FetchContext) CreateDepthHandler(node *meta.FetchNode, req *http.Request) *DepthHandler {
	f.CtxWait.Add(1)

	d := &DepthHandler{
		node: node,
	}

	d.Fetch = func() {
		defer f.CtxWait.Done()

		//优先判断广度
		if node.Event != nil && node.Event.Pageable != nil {
			f.WideChan <- f.CreateWideHandler(node, req)
			return
		}

		//获取页面元素
		doc, err := httpCall(req)
		if err != nil {
			log.Println(fmt.Sprintf("请求异常:[%s]:[%v]", req.URL.String(), err.Error()))
			return
		}
		//矩阵搜索
		f.MatrixChan <- f.CreateMatrixHandler(node, req, doc)
	}
	return d
}

func (f *FetchContext) CreateWideHandler(node *meta.FetchNode, req *http.Request) *WideHandler {
	f.CtxWait.Add(1)

	w := &WideHandler{
		node: node,
	}
	w.Fetch = func() {
		defer f.CtxWait.Done()

		pageWait := &sync.WaitGroup{}
		//轮询执行
		event := node.Event.Pageable
		startIndex := event.BeginIndex | 0
		for {
			//无最大限制，直到某一页没有数据返回，或者某页未抓取到数据
			if startIndex > event.EndIndex {
				break
			}
			pageWait.Add(1)

			copyNode := node.CopySelf()
			//将生成的节点添加同步节点
			node.AddChild(copyNode)

			//分页执行
			go func(idx int, sub *meta.FetchNode, pageWait *sync.WaitGroup) {
				defer pageWait.Done()

				// 广度分页创建请求时，将当前页的地址暴露给用户创建请求
				req := script.CreateRequest(sub, req, f.JsVm, event.FuncName, idx)
				if req == nil {
					return
				}
				//对每个节点设置请求路径
				sub.Referer = req.URL.String()

				//获取页面元素
				doc, err := httpCall(req)
				if err != nil {
					log.Println(fmt.Sprintf("请求异常:[%s]:[%v]", req.URL.String(), err.Error()))
					return
				}

				//对结果集添加到矩阵通道中，由矩阵处理
				f.MatrixChan <- f.CreateMatrixHandler(copyNode, req, doc)
			}(startIndex, copyNode, pageWait)

			//下一页
			startIndex++
		}

		pageWait.Wait()
	}
	return w
}

func (f *FetchContext) CreateMatrixHandler(node *meta.FetchNode, req *http.Request, dom *goquery.Document) *MatrixHandler {
	f.CtxWait.Add(1)

	m := &MatrixHandler{
		node: node,
	}
	m.Fetch = func() {
		defer f.CtxWait.Done()

		fetchArray := make([][]*meta.FetchData, 0)
		//定位
		selector := dom.Find(node.Bind.Position)

		log.Println("-------------------[" + node.Bind.Position + "]-------------------")
		log.Println(selector.Html())
		log.Println("-------------------[" + node.Bind.Position + "]-------------------")

		//遍历执行
		selector.Each(func(i int, selection *goquery.Selection) {
			fetchData := make([]*meta.FetchData, 0)
			//遍历执行
			for _, field := range node.Bind.Fields {

				fd := &meta.FetchData{
					Field: field.Mapper,
				}

				//格式化值类型
				switch field.ValueType {
				case meta.ArrayType: //数组
					array := make([]string, 0)
					//数组
					selection.Find(field.Selector).Each(func(s int, ss *goquery.Selection) {
						//传入nil字符，之前取当前节点相关数据
						data, err := findHandler("", field.FindType, ss)
						if err != nil {
							return
						}
						array = append(array, data)
					})
					fd.Value = array
				case meta.ObjectType: //对象
					out, err := findHandler(field.Selector, field.FindType, selection)
					if err != nil {
						continue
					}
					fd.Value = &struct {
						Name  string
						Value interface{}
					}{
						Name:  field.Mapper,
						Value: out,
					}
				default: //默认(值)
					out, err := findHandler(field.Selector, field.FindType, selection)
					if err != nil {
						continue
					}
					fd.Value = out
				}

				//添加到通道
				fetchData = append(fetchData, fd)
			}
			fetchArray = append(fetchArray, fetchData)
		})

		//判断如果当前结果集未null,则通知上级调用
		if len(fetchArray) == 0 {
			return
		}
		//添加值
		node.AppendData(fetchArray)

		//遍历
		f.eachCall(node, req, fetchArray)
	}
	return m
}

func (f *FetchContext) eachCall(currNode *meta.FetchNode, req *http.Request, fetchArray [][]*meta.FetchData) {

	event := currNode.Event
	//判断当前节点是否需要深度抓取
	if event == nil || event.Link == nil {
		return
	}

	//通过当前节点中的link.next指向的下一个规则判断
	depthNode := script.CreateLinkNode(f.Config.ScriptPath, event.Link.Next)

	//遍历进行深度抓取
	for _, v := range fetchArray {

		tmpNode := depthNode.CopySelf()
		currNode.AddChild(tmpNode)

		//构建每一个条目的请求
		req := script.CreateRequest(tmpNode, req, f.JsVm, event.Link.FuncName, v)
		if req == nil {
			continue
		}
		//对每个节点设置请求路径
		tmpNode.Referer = req.URL.String()
		//缓存上级值得来源
		tmpNode.From = v

		//创建深度实现
		f.DepthChan <- f.CreateDepthHandler(tmpNode, req)
	}
}

func httpCall(req *http.Request) (*goquery.Document, error) {

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}

	//请求
	client := &http.Client{Transport: transCfg, Timeout: time.Second * 180}

	//执行
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	read, err := charset.NewReader(resp.Body, resp.Header.Get("ContentType"))
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("请求[%v]", req.URL.String()))

	//解析dom
	dom, err := goquery.NewDocumentFromReader(read)
	if err != nil {
		return nil, err
	}

	return dom, nil
}

var findHandler = func(fs string, tp string, _s *goquery.Selection) (string, error) {
	if fs != "" {
		_s = _s.Find(fs)
	}

	//log.Println("-------------------[" + fs + "]-------------------")
	//log.Println(_s.Html())
	//log.Println("-------------------[" + fs + "]-------------------")

	//结果
	if tp == meta.FindText { //纯文本
		return _s.Text(), nil

	} else if tp == meta.FindHtml { //html
		return _s.Html()

	} else if strings.HasPrefix(tp, meta.FindAttr) { //属性
		out, bool := _s.Attr(strings.Split(tp, ":")[1])
		if !bool {
			return "", errors.New("not exist")
		}
		return out, nil

	}
	return "", errors.New("FindType not exist")
}
