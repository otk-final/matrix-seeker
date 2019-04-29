package meta

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
