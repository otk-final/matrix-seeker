package meta

import (
	"time"
)

type FetchConfig struct {
	HttpUrl    string
	ScriptPath string
	TimeOut    time.Duration
}

type FetchNode struct {
	Count     int        //子节点个数
	Level     int        //层级
	Bind      *NodeBind  `json:"bind"`
	Event     *NodeEvent `json:"event"`
	Data      [][]*FetchData
	Parent    *FetchNode
	Childrens []*FetchNode `json:"childrens"` //子节点
}

type FetchData struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}
