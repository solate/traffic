package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)


var uaList = []string{
	"Mozilla/5.0 (Linux; U; Android 2.2; en-gb; GT-P1000 Build/FROYO) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"Mozilla/5.0 (Linux; U; Android 4.0.4; en-gb; GT-I9300 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
	"Mozilla/5.0 (Linux; Android 4.1.1; Nexus 7 Build/JRO03D) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.166  Safari/535.19",
	"Mozilla/5.0 (Android; Mobile; rv:14.0) Gecko/14.0 Firefox/14.0",
	"Mozilla/5.0 (Android; Tablet; rv:14.0) Gecko/14.0 Firefox/14.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:21.0) Gecko/20100101 Firefox/21.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:21.0) Gecko/20130331 Firefox/21.0",
	"Mozilla/5.0 (Windows NT 6.2; WOW64; rv:21.0) Gecko/20100101 Firefox/21.0",
}

//资源
type Resource struct {
	URL string //URL
	Target string //来源
	Start int //开始
	End int //结束
}

func RuleResource() (res []Resource) {

	//首页
	r1 := Resource{
		URL: "http://localhost:8080/",
		Target: "",
		Start: 0,
		End: 0,
	}

	//列表页
	r2 := Resource{
		URL: "http://localhost:8080/list/{$id}.html",
		Target: "{$id}",
		Start: 1,
		End: 21,
	}

	//详情页
	r3  := Resource{
		URL: "http://localhost:8080/detail/{$id}.html",
		Target: "{$id}",
		Start: 1,
		End: 12924,
	}

	res = append(res, r1, r2, r3)

	return
}


func main() {
	//生成日志行数
	total := flag.Int("total", 100, "how many rows by created")
	//生成文件目标地址
	filePath := flag.String("filePath", "/Users/jinzhang/workspace/nginx/logs/dig.log", "log file path")
	flag.Parse()

	fmt.Println("total: ",*total," , filePath: ", *filePath)

	// 构造网站真实url集合
	res := RuleResource()
	list := BuildURL(res)

fmt.Println(list)

	var logStr string
	// 生成total行日志内容，使用上面的集合
	for i:=0; i<=*total ; i++  {

		current := list[randInt(0, len(list)-1)]
		refer := list[randInt(0, len(list)-1)]
		ua := uaList[randInt(0, len(uaList)-1)]
		logStr += MakeLog(current, refer, ua) + "\n"
		fmt.Println(logStr)
	}

	fd, _ := os.OpenFile(*filePath, os.O_RDWR| os.O_APPEND, 0644)
	defer fd.Close()
	fd.WriteString(logStr)



	fmt.Println("------END-----")

}

//随机数
func randInt(min int, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min > max {
		return max
	}
	return r.Intn(max-min) + min
}

//创建日志
func MakeLog(current, refer, ua string) (str string) {
	u := url.Values{}
	u.Set("time", "1")
	u.Set("url", current)
	u.Set("refer", refer)
	u.Set("ua", ua)
	paramStr := u.Encode()

	template := `127.0.0.1 - - [14/Nov/2018:15:57:35 +0800] "GET /dig?%s HTTP/1.1" 200 43 %s "-"`
	str = fmt.Sprintf(template, paramStr, ua)
	return
}

//生成url
func BuildURL(res []Resource) (list []string) {

	for _,v := range res {
		if len(v.Target) == 0 {
			list = append(list, v.URL)
		}else {
			for i:=v.Start; i<=v.End; i++  {
				urlString := strings.Replace(v.URL, v.Target, strconv.Itoa(i), -1)
				list = append(list, urlString)
			}
		}
	}

	return
}