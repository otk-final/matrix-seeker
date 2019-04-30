var Host = "https://www.meixueyuan.com/";

/**
 初始化根请求
 */
function startRoot(node, formReq, arg) {
    var nextReq = {
        url: Host,
        method: "GET"
    };
    return nextReq;
}

//首页分类
function specialClassifyLink(node, fromReq, arg) {

    //第一个为最新活动，没有地址，所以返回nil
    var curl = arg["classifyUrl"];
    if (!curl) {
        return null
    }
    var nextReq = {
        url: Host + curl,
        method: "GET"
    };
    return nextReq;
}


//专区分类
function specialClassifyListLink(node, fromReq, arg) {
    var nextReq = {
        url: Host + arg["specialUrl"],
        method: "GET"
    };
    return nextReq;
}

//
function specialClassifyListPage(node, fromReq, arg) {
    var fromUrl = fromReq.url;
    console.info("fromUrl:" + fromUrl);
}