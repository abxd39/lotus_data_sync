package utils

import (
	"fmt"
	"github.com/arthurkiller/rollingwriter"
	"os"
	"strings"
	"sync"
)

var file *os.File
var mutx sync.Mutex
var Log *Logger

var MongodbLog *Logger
var SqlLoger *Logger

func init() {
	var err error
	file, err = os.OpenFile("./logger/ss.logger", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
}

func SetupLogger() error {

	config := rollingwriter.Config{
		LogPath:                "../xorm-logs",              //日志路径
		TimeTagFormat:          "060102150405",              //时间格式串
		FileName:               "mysql_xorm",                //日志文件名
		MaxRemain:              3,                           //配置日志最大存留数
		RollingPolicy:          rollingwriter.VolumeRolling, //配置滚动策略 norolling timerolling volumerolling
		RollingTimePattern:     "* * * * * *",               //配置时间滚动策略
		RollingVolumeSize:      "500M",                      //配置截断文件下限大小
		WriterMode:             "none",
		BufferWriterThershould: 256,
		// Compress will compress log file with gzip
		Compress: true,
	}

	var err error
	writer, err := rollingwriter.NewWriterFromConfig(&config)
	if err != nil {
		panic(err)
	}
	_ = writer
	logfilePath := Initconf.String("logFile")
	//if logfilePath == "" {
	//	panic("logfilePath 为空")
	//}

	Log, err = NewLogger(logfilePath, 1)
	if err != nil {
		panic(err)
	}
	mongodbLog := Initconf.String("mongoLog")
	MongodbLog, err = NewMongoDbLogger(mongodbLog, 1)
	if err != nil {
		panic(err)
	}

	return nil
}

// TODO: use log4go to replace this sample logger..
func Printf(prefix string, fmts string, args ...interface{}) {
	mutx.Lock()
	defer mutx.Unlock()

	if prefix = strings.Trim(prefix, " "); prefix != "" {
		fmts = "%s:" + fmts
		args = append([]interface{}{prefix}, args[:]...)
	}

	if l := len(fmts); fmts[l-1] != '\n' {
		fmts += "\n"
	}

	message := fmt.Sprintf(fmts, args[:]...)
	fmt.Printf(message)
	if file != nil {
		fmt.Fprintf(file, message)
	}
}
