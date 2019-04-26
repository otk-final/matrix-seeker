package meta

import (
	"encoding/json"
	"fmt"
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
	Position string           `json:"position"`
	Fields   [] *NodePosition `json:"fields"`
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
	Selector  string    `json:"selector"`  //选择器
	FindType  FindType  `json:"findType"`  //查找类型
	ValueType ValueType `json:"valueType"` //值类型
	Mapper    string    `json:"mapper"`    //映射名称
}

func (node *FetchNode) CopySelf() *FetchNode {
	return node
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
	var parent FetchNode

	parent.Count = parent.Count + len(siblings)
}
