package main

import (
	"bufio"
	"flag"
	"github.com/labstack/gommon/log"
	"io"
	"os"

	"time"
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
func LogConsumer(strings chan string, data chan UrlData, data2 chan UrlData) {

	

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

