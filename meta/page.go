package meta

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FindType = string
type ValueType = string

const (
	FindText   FindType  = "text" //文本
	FindHtml   FindType  = "html" //html文档
	FindAttr   FindType  = "attr" //属性
	ArrayType  ValueType = "[]"   //数组
	TextType   ValueType = "_"    //字段
	ObjectType ValueType = "{}"   //对象
)

type NodeBind struct {
	Position string          `json:"position"`
	Fields   []*NodePosition `json:"fields"`
}

type NodeEvent struct {
	Link     *LinkEvent     `json:"link"`     //跳转
	Pageable *PageableEvent `json:"pageable"` //分页
}

type LinkEvent struct {
	FuncName string `json:"funcName"` //函数名
	Next     string `json:"next"`
}

type PageableEvent struct {
	FuncName   string `json:"funcName"`   //函数名
	BeginIndex int    `json:"beginIndex"` //开始页
	EndIndex   int    `json:"endIndex"`   //结束页
}

type NodePosition struct {
	Selector   string    `json:"selector"`   //选择器
	FindType   FindType  `json:"findType"`   //查找类型
	ValueType  ValueType `json:"valueType"`  //值类型
	ActionType string    `json:"actionType"` //动作类型
	Mapper     string    `json:"mapper"`     //映射名称
}

func (node *FetchNode) CopySelf() *FetchNode {
	return &FetchNode{
		Count:     node.Count,
		Level:     node.Level,
		Bind:      node.Bind,
		Event:     node.Event,
		Data:      make([][]*FetchData, 0),
		Childrens: make([]*FetchNode, 0),
	}
}

func (node *FetchNode) AppendData(temp [][]*FetchData) {
	node.Data = append(node.Data, temp...)

	by, _ := json.Marshal(temp)
	fmt.Println(string(by))
}

/*
	添加子节点
*/
func (node *FetchNode) AddChild(subs ...*FetchNode) {
	//修改数量
	node.Count = node.Count + len(subs)
	//修改层级
	for _, sub := range subs {
		sub.Level = node.Level + 1
	}
	node.Childrens = append(node.Childrens, subs...)
}

/*
	添加同胞节点
*/
func (node *FetchNode) AddSiblings(siblings ...*FetchNode) {
	for _, sib := range siblings {
		sib.Level = node.Level
	}

	//父节点
	parent := node.Parent
	if parent != nil {
		parent.Childrens = append(parent.Childrens, siblings...)
		parent.Count = parent.Count + len(siblings)
	} else {
		//root
		node.AddChild(siblings...)
	}
}

func (node *FetchNode) GetNodeFilePath(rootDir string) string {

	loopParent := func(names []string, cur *FetchNode) ([]string, *FetchNode) {
		return append([]string{"node" + strconv.Itoa(cur.Level)}, names[:]...), cur.Parent
	}

	out := make([]string, 0)

	//递归查询
	parent := node
	for {
		out, parent = loopParent(out, parent)
		if parent == nil || parent.Level == 0 {
			break
		}
	}

	return strings.Join(out, "/")
}
