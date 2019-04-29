package artifact

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/*
	下载图片
*/
func (at *Persistent) CreateImgTask(fileDir string, referer string, imgUrl string) {
	defer at.WaitGroup.Done()

	u, err := url.Parse(imgUrl)
	if err != nil {
		return
	}

	//去掉最左边的'/'
	tmp := strings.TrimLeft(u.Path, "/")
	imgPath := fileDir + strings.ToLower(strings.Replace(tmp, "/", "-", -1))

	//检查文件是否存在
	if imgFile, _ := os.Stat(imgPath); imgFile != nil {
		return
	}

	//请求
	log.Println(fmt.Sprintf("下载img[%s]", imgUrl))

	/*
		构建请求
	*/
	reqUrl, _ := url.Parse(imgUrl)
	req := &http.Request{
		Header: http.Header{
			"referer": []string{referer},
		},
		URL:    reqUrl,
		Method: "GET",
	}
	//调用(超时1分钟）
	client := http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	imgData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	//写入文件
	waitToFile(imgPath, imgData)
}
