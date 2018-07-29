package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
	"io"
	"time"
)

type LogProcess struct {
	rc    chan []byte // read->process
	wc    chan string // process->write
	read  Reader      // 日志读取模块
	write Writer      // 写入模块
}

func (l *LogProcess) Process() {
	// 解析模块
	for v := range l.rc {
		l.wc <- strings.ToUpper(string(v))
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
	Write(wc chan string)
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

func (w *WriteToInfluxDB) Write(wc chan string) {
	for v := range wc {
		fmt.Println(v)
	}
}
