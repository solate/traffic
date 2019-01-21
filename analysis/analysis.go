package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"github.com/labstack/gommon/log"
	"io"
	"net/url"
	"os"
	"strings"

	"time"
)

const (
	Handle_Dig = " /dig?"
	Handle_End = " HTTP/"
)

//传入参数
type CmdParams struct {
	LogFilePath string
	RoutineNum  int
}

//消费数据
type UrlData struct {
	Data DigData
	Uid string

}
//日志数据
type DigData struct {
	Time string //时间
	URL string //路径
	Refer string //上一级
	Ua string //客户端
}

//存储结构
type StorageBlock struct {
	CounterType string //存储统计类型pv/uv
	storageModel string //存储格式
	UrlNode UrlNode
}
type UrlNode struct {
	//
}

//var log = logrus.New()
//func init()  {
//	log.Out = os.Stdout
//	log.SetLevel(logrus.DebugLevel)
//
//}

func main() {

	// # 获取参数
	// ## 读取日志位置
	logFilePath := flag.String("logFilePath", "/Users/jinzhang/workspace/nginx/logs/dig.log", "log file path")
	// ## goroutine并发数，开启多少并发进行分析
	routineNum := flag.Int("routineNum", 5 , "consumer goroutine num")
	// 这个项目打印的运行日志输出到哪里
	targetLogFilePath := flag.String("l", "log.txt", "project runtime log  file path")
	flag.Parse()

	params := CmdParams{
		LogFilePath: *logFilePath,
		RoutineNum: *routineNum,
	}

	//打日志
	logFd, err := os.OpenFile(*targetLogFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.SetOutput(logFd)
		defer logFd.Close()
	}
	log.Info("Exec start.")
	log.Info("Params:", params)


	//初始化channel,用于数据传递
	logChannel := make(chan  string, 3*(*routineNum)) //日志
	pvChannel := make(chan UrlData, *routineNum)  //pv
	uvChannel := make(chan UrlData, *routineNum)  //uv
	storageChannel := make(chan StorageBlock, *routineNum) //存储


	//日志消费者
	go ReadFileByLine(params,logChannel)

	//创建一组日志处理
	for i:=0; i<params.RoutineNum; i++  {
		go LogConsumer(logChannel, pvChannel, uvChannel)
	}

	//创建PV UV 统计器
	go PvCounter(pvChannel, storageChannel)
	go UVCounter(uvChannel, storageChannel)

	//创建存储器
	go DataStorage(storageChannel)


	time.Sleep(1000*time.Second)
}




//按行消费日志文件
func ReadFileByLine(params CmdParams, logChannel chan string) (err error) {

	fd, err := os.Open(params.LogFilePath)
	if err != nil {
		log.Warnf("ReadFileByLine can't open file:%s", params.LogFilePath)
		return err
	}
	defer  fd.Close()

	count := 0 //计数器
	bufferReader := bufio.NewReader(fd)
	for  {
		line, _, err := bufferReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Infof("ReadFileByLine wait, read line: %d", count)
				time.Sleep(3 * time.Second) //如果读到日志末尾，那么休息3秒
			}

			log.Warnf("ReadFileByLine read log error")
			//return err
		}
		logChannel <- string(line)
		count++

		if count % (1000 * params.RoutineNum) == 0 {
			log.Infof("ReadFileByLine line: %d", count)
		}

	}


	return
}

//日志处理
func LogConsumer(logChannel chan string, pvChannel chan UrlData, uvChannel chan UrlData) (err error) {

	//逐行消费
	for logStr := range logChannel {
		// 切割日志字符串，得到打点上报的日志
		data, err := cuteLogFetchData(logStr)
		if err != nil {
			return err
		}

		//uid : 模拟生成uid md5(refer + ua)

		hasher := md5.New()
		hasher.Write([]byte(data.Refer + data.Ua))
		uid := hex.EncodeToString(hasher.Sum(nil)) //16进制字符串


		//解析都可以放这里
		uData := UrlData{ data, uid}


		log.Info("uData : ==>", uData)

		pvChannel <- uData
		uvChannel <- uData




	}

	return

}

//切割字符串,获得打点数据
func cuteLogFetchData(logStr string) (dig DigData, err error) {

	logStr = strings.TrimSpace(logStr)
	pos1 := strings.Index(logStr, Handle_Dig)
	if pos1 == -1 { //没找到
		return
	}
	pos1 += len(Handle_Dig)//位置 + 偏移

	pos2 := strings.Index(logStr, Handle_End)

	d := logStr[pos1: pos2]

	urlInfo, err := url.Parse("http://localhost?" +d)
	if err != nil {
		return
	}


	data := urlInfo.Query()

	dig = DigData{
		data.Get("time"),
		data.Get("url"),
		data.Get("refer"),
		data.Get("ua"),
	}







	return
}

//pv统计
func PvCounter(data chan UrlData, blocks chan StorageBlock) {

}

//uv统计
func UVCounter(data chan UrlData, blocks chan StorageBlock) {

}

//数据存储
func DataStorage(blocks chan StorageBlock) {

}

