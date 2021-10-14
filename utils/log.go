package utils

import (
	"sync"
)

//var file *os.File
var mutx sync.Mutex
var Log *Logger

var MongodbLog *Logger
var SqlLoger *Logger

// func init() {
// 	var err error
// 	file, err = os.OpenFile("./logger/ss.logger", os.O_CREATE|os.O_WRONLY, 0777)
// 	if err != nil {
// 		fmt.Println("open file error:", err)
// 		return
// 	}
// }

func SetupLogger() error {

	logfilePath := Initconf.String("logFile")
	//if logfilePath == "" {
	//	panic("logfilePath 为空")
	//}
	var err error
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
// func Printf(prefix string, fmts string, args ...interface{}) {
// 	mutx.Lock()
// 	defer mutx.Unlock()

// 	if prefix = strings.Trim(prefix, " "); prefix != "" {
// 		fmts = "%s:" + fmts
// 		args = append([]interface{}{prefix}, args[:]...)
// 	}

// 	if l := len(fmts); fmts[l-1] != '\n' {
// 		fmts += "\n"
// 	}

// 	message := fmt.Sprintf(fmts, args[:]...)
// 	fmt.Printf(message)
// 	if file != nil {
// 		fmt.Fprintf(file, message)
// 	}
// }
