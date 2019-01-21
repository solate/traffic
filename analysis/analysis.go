package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"

	"github.com/labstack/gommon/log"

	"time"
)

const (
	Handle_Dig = " /dig?"
	Handle_End = " HTTP/"

	Handle_Movie = "/movie/"
	Handle_List  = "/list/"
	Handle_Html  = ".html"
)

//传入参数
type CmdParams struct {
	LogFilePath string
	RoutineNum  int
}

//消费数据
type UrlData struct {
	Data    DigData
	Uid     string
	UrlNode UrlNode
}

//日志数据
type DigData struct {
	Time  string //时间
	URL   string //路径
	Refer string //上一级
	Ua    string //客户端
}

//存储结构
type StorageBlock struct {
	CounterType  string //存储统计类型pv/uv
	storageModel string //存储格式
	UrlNode      UrlNode
}
type UrlNode struct {
	UrlType       string //url类型, /movie/ 或 /list/ ...
	UrlResourceID int    //url 资源id
	Url           string //url 当前页面url
	UrlTime       string //当前访问这个页面的时间
}

//var log = logrus.New()
//func init()  {
//	log.Out = os.Stdout
//	log.SetLevel(logrus.DebugLevel)
//
//}

var redisClient *redis.Pool

func init() {

	// 建立连接池, 这里正式项目需要更完善一些
	redisClient = &redis.Pool{
		MaxIdle:     5,
		MaxActive:   5,
		IdleTimeout: 10 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}
}

func main() {

	// # 获取参数
	// ## 读取日志位置
	logFilePath := flag.String("logFilePath", "/Users/jinzhang/workspace/nginx/logs/dig.log", "log file path")
	// ## goroutine并发数，开启多少并发进行分析
	routineNum := flag.Int("routineNum", 5, "consumer goroutine num")
	// 这个项目打印的运行日志输出到哪里
	targetLogFilePath := flag.String("l", "log.txt", "project runtime log  file path")
	flag.Parse()

	params := CmdParams{
		LogFilePath: *logFilePath,
		RoutineNum:  *routineNum,
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
	logChannel := make(chan string, 3*(*routineNum))       //日志
	pvChannel := make(chan UrlData, *routineNum)           //pv
	uvChannel := make(chan UrlData, *routineNum)           //uv
	storageChannel := make(chan StorageBlock, *routineNum) //存储

	//日志消费者
	go ReadFileByLine(params, logChannel)

	//创建一组日志处理
	for i := 0; i < params.RoutineNum; i++ {
		go LogConsumer(logChannel, pvChannel, uvChannel)
	}

	//创建PV UV 统计器
	go PvCounter(pvChannel, storageChannel)
	go UVCounter(uvChannel, storageChannel)

	//创建存储器
	go DataStorage(storageChannel)

	time.Sleep(1000 * time.Second)
}

//按行消费日志文件
func ReadFileByLine(params CmdParams, logChannel chan string) (err error) {

	fd, err := os.Open(params.LogFilePath)
	if err != nil {
		log.Warnf("ReadFileByLine can't open file:%s", params.LogFilePath)
		return err
	}
	defer fd.Close()

	count := 0 //计数器
	bufferReader := bufio.NewReader(fd)
	for {
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

		if count%(1000*params.RoutineNum) == 0 {
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
		uData := UrlData{data, uid, formartUrl(data.URL, data.Time)}

		pvChannel <- uData
		uvChannel <- uData

	}

	return

}

//格式化url => UrlNode
func formartUrl(url string, t string) (data UrlNode) {

	//详情页
	pos1 := strings.Index(url, Handle_Movie)
	if pos1 != -1 {
		pos1 += len(Handle_Movie)

		pos2 := strings.Index(url, Handle_Html)
		idStr := url[pos1:pos2]

		id, _ := strconv.Atoi(idStr)

		data = UrlNode{"Movie", id, url, t}
	}

	//列表页
	pos1 = strings.Index(url, Handle_List)
	if pos1 != -1 {
		pos1 += len(Handle_List)

		pos2 := strings.Index(url, Handle_Html)
		idStr := url[pos1:pos2]

		id, _ := strconv.Atoi(idStr)

		data = UrlNode{"List", id, url, t}
	}

	//首页
	data = UrlNode{"Index", 0, url, t}

	return
}

//切割字符串,获得打点数据
func cuteLogFetchData(logStr string) (dig DigData, err error) {

	logStr = strings.TrimSpace(logStr)
	pos1 := strings.Index(logStr, Handle_Dig)
	if pos1 == -1 { //没找到
		return
	}
	pos1 += len(Handle_Dig) //位置 + 偏移

	pos2 := strings.Index(logStr, Handle_End)

	d := logStr[pos1:pos2]

	urlInfo, err := url.Parse("http://localhost?" + d)
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

//pv统计, 每天访问量
func PvCounter(pvChannel chan UrlData, storageChannel chan StorageBlock) {

	for data := range pvChannel {
		sItem := StorageBlock{"pv", "ZINCREBY", data.UrlNode}
		storageChannel <- sItem
	}

}

//uv统计, 每个用户每天访问量
func UVCounter(uvChannel chan UrlData, storageChannel chan StorageBlock) (err error) {

	// 从池里获取连接
	rc := redisClient.Get()
	// 用完后将连接放回连接池
	defer rc.Close()

	//需要去重 HyperLoglog redis 内置的数据类型 pv>=uv
	for data := range uvChannel {

		//去重
		hyperLogLogKey := "uv_hpll_" + GetTime(data.Data.Time, "day")
		ret, err := redis.Int(rc.Do("PFADD", hyperLogLogKey, data.Uid, "EX", 86400))
		if err != nil {
			log.Warnf("UVCounter check HyperLoglog failed: %s", err.Error()) //错了当没有
		}
		if ret != 1 { //已存在
			continue
		}

		sItem := StorageBlock{"uv", "ZINCREBY", data.UrlNode}
		storageChannel <- sItem
	}

	return

}

//在一个时间段内去重
func GetTime(logTime, timeType string) (timeStr string) {

	var item string
	switch timeType {
	case "day": //天级别
		item = "2006-01-02"
	case "hour ":
		item = "2006-01-02 15"
	case "min":
		item = "2006-01-02 15:04"

	}

	t, _ := time.Parse(logTime, time.Now().Format(item))

	timeStr = strconv.FormatInt(t.Unix(), 10)
	return
}

//数据存储
func DataStorage(blocks chan StorageBlock) {
	//存redis ,这里要做连接池,这个测试使用一下

}
