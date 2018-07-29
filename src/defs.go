package main

import (
	"fmt"
	"os"
	"bufio"
	"io"
	"time"
	"regexp"
	"log"
	"strconv"
	"strings"
	"net/url"
)

type LogProcess struct {
	rc    chan []byte   // read->process
	wc    chan *Message // process->write
	read  Reader        // 日志读取模块
	write Writer        // 写入模块
}

// 存储提取的监控数据
type Message struct {
	TimeLocal                    time.Time
	BytesSent                    int
	Path, Method, Scheme, Status string
	UpstreamTime, RequestTime    float64
}

func (l *LogProcess) Process() {
	// 解析模块
	// 从rc读取每一行日志数据->正则提取所需的监控数据->写入wc
	// 172.0.0.12 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	r := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)

	location, _ := time.LoadLocation("Asia/Shanghai")

	for v := range l.rc {
		ret := r.FindStringSubmatch(string(v))
		if len(ret) != 14 {
			log.Println("FindStringSubmatch fail:", string(v))
		}

		msg := &Message{}
		t, err := time.ParseInLocation("02/Jan/2006:15:04:05 +0000", ret[4], location)
		if err != nil {
			log.Println("ParseInLocation fail:", err.Error(), ret[4])
			continue
		}
		msg.TimeLocal = t

		byteSent, _ := strconv.Atoi(ret[8])
		msg.BytesSent = byteSent

		// GET /foo?query=t HTTP/1.0
		reqSli := strings.Split(ret[6], " ")
		if len(reqSli) != 3 {
			log.Println("strings.Split fail", ret[6])
			continue
		}
		msg.Method = reqSli[0]

		u, err := url.Parse(reqSli[1])
		if err != nil {
			log.Println("url parse fail:", err)
			continue
		}
		msg.Path = u.Path

		msg.Scheme = ret[5]
		msg.Status = ret[7]

		upstreamTime, _ := strconv.ParseFloat(ret[12], 64)
		requestTime, _ := strconv.ParseFloat(ret[13], 64)
		msg.UpstreamTime = upstreamTime
		msg.RequestTime = requestTime

		l.wc <- msg
	}
}

type ReadFromFile struct {
	path string // 读取文件的路径
}

type WriteToInfluxDB struct {
	influxDBDsn string // influx data source
}

// 将读写模块抽象为借口
type Reader interface {
	Read(rc chan []byte)
}

type Writer interface {
	Write(wc chan *Message)
}

func (r *ReadFromFile) Read(rc chan []byte) {
	// 从文件读取
	// 打开文件
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("Open file error: %s", err.Error()))
	}

	// 字符指针移动到文件末尾开始逐行读取文件内容
	f.Seek(0, 2)
	rd := bufio.NewReader(f)
	// 读取直到换行符

	for {
		line, err := rd.ReadBytes('\n')
		// 读取到文件末尾
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("Read bytes error: %s", err.Error()))
		}
		rc <- line[:len(line)-1]
	}
}

func (w *WriteToInfluxDB) Write(wc chan *Message) {
	for v := range wc {
		fmt.Println(v)
	}
}
