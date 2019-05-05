var Host = "https://www.taxueai.com/";


/**
 初始化根请求
 */
function startRoot(node, formReq, arg) {
    var nextReq = {
        url: Host + "qgal",
        method: "GET"
    }
    return nextReq;
}

//分类首页
function page1(node, formReq, arg) {
    var nextReq = {
        url: arg["classifyUrl"],
        method: "GET"
    };
    return nextReq;
}

//跳转详情页
function page2(node, formReq, arg) {
    var nextReq = {
        url: arg["groupUrl"],
        method: "GET"
    };
    return nextReq;
}

//列表分页
function page2ForPageable(node, formReq, arg) {
    var nextPage = formReq['url']+'/page_'+arg+'.html';
    console.info(nextPage);
    var nextReq = {
        url: nextPage,
        method: "GET"
    };
    return nextReq;
}


//分类分页
function page3(node, formReq, arg) {
    var nextReq = {
        url: arg['contentUrl'],
        method: "GET"
    };
    return nextReq;
}
