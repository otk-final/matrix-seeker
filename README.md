# matrix-seeker
矩阵爬虫
### 设计
![原型](https://github.com/otk-final/matrix-seeker/tree/master/script-example/img/design.png)
* 深度 **Depth**  
 请求数据 > 生成dom > **矩阵**
--- 
* 广度 **Wide**  
 分页执行 > 生成多个分页请求 > **深度**
---
* 矩阵 **Matrix**  
 定位dom > 解析bind > 存储数据 > 生成跳转请求 > **广度**
---  
> 注意：每个节点event配置属性决定分支流转的走向  
> 当前节点如果配置 event.pageable 则构建广度模型  
> 当前节点如果配置 event.link 则执行深度模型  
> 均配置优先广度模型

### 配置(脚本)
1.页面节点 (自定义.json)   
2.请求构建 (script.js)  
所有脚本文件需放在同一目录下,**必须有含有一个(root.json)作为根节点** 

**页面定位**  
![示例](https://github.com/otk-final/matrix-seeker/tree/master/script-example/img/nodeJson.png)

> name 当前节点名称（用户输出到同一文件目录）  
> bind.position  当前页面定位  
> bind.fields  元素集  
>> field.selector  goquery选择器语法 (基于position子节点)
>> field.findType  查找类型（[text]:文本,[html]:内容,[attr:?]属性    
>> field.actionType  download 如果当前抓取的值是图片地址，需下载图片则标记为download
>> field.valueType  值类型 '[],数组','_,默认','{},对象'  
>> field.mapper  映射标识
>  
> event.link 跳转事件  
>> link.funcName 生成跳转请求函数  
>> link.next 下一个新的节点页面
>    
> event.pageable 分页事件  
>> pageable.funcName 生成分页请求函数  
>> pageable.beginIndex 开始页码  
>> pageable.endIndex 结束页码  
>

* 定位文件必须为有效json文件, 根据当前页面分页(广度)需要，和跳转(深度)抓取需要配置event事件，无则不进行设置  

**请求构建**

* 必须含有方法 ``function startRoot(node, formReq, arg)`` 用户初始化请求

* 其他方法均分为跳转函数，分页函数，自定义函数

类型 |名称 | node | formReq | arg
:---|:---|:---|:---|:---
初始化|startRoot|root.json|nil|nil
跳转|见页面.json文件中link.funcName|页面.json|当前页面请求源|bind.fields中抓取的数据
分页|见页面.json文件中pageable.funcName|页面.json|当前页面请求源|页码

fromReq与函数返回的req请求格式一致，建议值类型统一采用string类型，防止参数不兼容
```javascript
let fromReq = {
	url:"http://www.xxx.com",
	method:"GET|POST|...",
	params:{
		param1:"value1",
		param2:"value2"
	},
	header:{
		user:"111",
		token:"f02f6989-5627-4493-a153-47621591e1eb"
	}
}
```

* 跳转函数中arg参数均转换为json结构 即原始数据为
```json
[
    {
      "field": "classifyUrl",
      "value": "/qinggan/lianai/"
    },
    {
      "field": "classifyName",
      "value": "谈恋爱技巧"
    }
  ]
```
因为js不太方便获取相关值，均将json数组转换成json对象形式,所以定义页面节点文件时,mapper不要重复
```json
    {
      "classifyUrl": "/qinggan/lianai/",
      "classifyName": "谈恋爱技巧"
    }
```


### 运行

* 结果文件在当前脚本文件夹下生产out目录,已节点名称为二级目录,结果文件均为json格式  
文件名：``C:\Users\KF\Desktop\项目\out\节点名称\e71ca970-b77b-4269-a179-cfae68b8ce1c.json``  
```json
[
  [
    {
      "field": "classifyUrl",
      "value": "/qinggan/lianai/"
    },
    {
      "field": "classifyName",
      "value": "谈恋爱技巧"
    }
  ],
  [
    {
      "field": "classifyUrl",
      "value": "/qinggan/fqgx/"
    },
    {
      "field": "classifyName",
      "value": "夫妻关系"
    }
  ]
]
```

* 所有请求默认超时为1分钟
* 同一节点下相同图片地址，不做重复下载，因为各大网站对图片素材都做了响应防爬措施,会有下载失败场景
* 爬取日志文件位于当前项目路径out/seeker.log

执行matrix-seeker.exe 执行(window64位)如下
![开始](https://github.com/otk-final/matrix-seeker/tree/master/script-example/img/start.png)
![结束](https://github.com/otk-final/matrix-seeker/tree/master/script-example/img/finish.png)

## Feature

* 对script.js支持进行外部js导入
* 目前只支持本地机器爬取，所以任务请求均并发处理，未对请求间隔做控制  
* 暂不支持代理模式
