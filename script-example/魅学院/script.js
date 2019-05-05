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

    var path = fromUrl.substring(fromUrl.lastIndexOf('/') + 1);
    var pathArray = path.split('-');
    var nextPath = pathArray[0] + "-" + pathArray[1] + "-" + arg + ".html";

    console.info("nextUrl:" + Host + nextPath);

    var nextReq = {
        url: Host + nextPath,
        method: "GET"
    };
    return nextReq;
}

//详情页
function contentLink(node, fromReq, arg) {
    var contentUrl = arg['contentUrl'];

    var nextReq = {
        url: Host + contentUrl,
        method: "GET"
    };
    return nextReq;
}