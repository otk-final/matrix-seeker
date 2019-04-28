var Host = "http://www.7y7.com";


/**
 初始化根请求
 */
function startRoot(node, formReq, arg) {
    var nextReq = {
        url: Host + "/qinggan",
        method: "GET"
    }
    return nextReq;
}

//分类首页
function builderClassifyHome(node, formReq, arg) {
    var nextReq = {
        url: Host + arg["classifyUrl"],
        method: "GET"
    }
    return nextReq;
}

//分类分页
function builderClassifyPage(node, formReq, arg) {
    var url = formReq["url"] + "index_" + arg + ".html";
    var nextReq = {
        url: url,
        method: "GET"
    }
    return nextReq;
}

//内容详情
function builderContentPage(node, formReq, arg) {

    var url = Host + arg["detailUrl"];

    var nextReq = {
        url: url,
        method: "GET"
    }

    return nextReq;
}