package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type FetchConfig struct {
	ScriptPath string
	Interval   time.Duration
	TimeOut    time.Duration
}

type FetchNode struct {
	Count     int //子节点个数
	Level     int //层级
	Referer   string
	Name      string     `json:"name"` //名称
	Bind      *NodeBind  `json:"bind"`
	Event     *NodeEvent `json:"event"`
	Data      [][]*FetchData
	From      []*FetchData
	Childrens []*FetchNode `json:"childrens"` //子节点
}

type FetchData struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

type FileFetchData struct {
	Referer string         `json:"referer"`
	Data    [][]*FetchData `json:"data"`
	From    []*FetchData   `json:"from"`
}

func (node *FetchNode) CopySelf() *FetchNode {
	return &FetchNode{
		Referer:   node.Referer,
		Name:      node.Name,
		Count:     node.Count,
		Level:     node.Level,
		Bind:      node.Bind,
		Event:     node.Event,
		From:      node.From,
		Data:      make([][]*FetchData, 0),
		Childrens: make([]*FetchNode, 0),
	}
}

func (node *FetchNode) AppendData(temp [][]*FetchData) {
	node.Data = append(node.Data, temp...)

	by, _ := json.Marshal(temp)
	log.Println(fmt.Sprintf("数据:%s", by))
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

//获取字段
func (node *FetchNode) getActionFields(actionType string) map[string]interface{} {
	fields := make(map[string]interface{}, 0)
	for _, field := range node.Bind.Fields {
		if field.ActionType == actionType {
			fields[field.Mapper] = "ok"
		}
	}
	return fields
}

func (node *FetchNode) GetActionValues(actionType string) []string {

	fields := node.getActionFields(actionType)

	dataEach := func(item []*FetchData) []string {
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

	outs := make([]string, 0)
	for _, item := range node.Data {
		outs = append(outs, dataEach(item)...)
	}

	return outs
}
