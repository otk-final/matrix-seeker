package artifact

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/*
	下载图片
*/
func (at *Persistent) CreateImgTask(imgUrl string, fileDir string) {

	u, err := url.Parse(imgUrl)
	if err != nil {
		return
	}

	//去掉最左边的'/'
	tmp := strings.TrimLeft(u.Path, "/")
	imgPath := fileDir + strings.ToLower(strings.Replace(tmp, "/", "-", -1))

	//检查文件是否存在
	if _, err := os.Stat(imgPath); os.IsExist(err) {
		return
	}

	//请求
	resp, err := http.Get(imgUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	imgData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	imgFile, err := os.Create(imgPath)
	if err != nil {
		return
	}
	defer imgFile.Close()

	//写入文件
	imgFile.Write(imgData)
}
