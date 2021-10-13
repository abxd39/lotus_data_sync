/**
 * @Author: wangyingwen
 * @Description:
 * @File:  log_new
 * @Version: 1.0.0
 * @Date: 2021/7/20 下午1:56
 */

package utils

import (
	"fmt"
	"io"
	"log"
	"os"

)

type Logger struct {
	LogFile    string
	TraceLevel int
	trace      *log.Logger
	info       *log.Logger
	warn       *log.Logger
	error      *log.Logger
}

//var Mongolog *Logger
//var ProjectLog *Logger

func NewLogger(logfile string, tracelevel int) (*Logger, error) {
	ProjectLog := new(Logger)
	ProjectLog.LogFile = logfile
	ProjectLog.TraceLevel = tracelevel
	if w, err := ProjectLog.getWriter(); err != nil {
		return ProjectLog, err
	} else {
		ProjectLog.trace = log.New(w, "[filscan---T] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		ProjectLog.info = log.New(w, "[filscan---I] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		ProjectLog.warn = log.New(w, "[filscan---W] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		ProjectLog.error = log.New(w, "[filscan---E] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
		return ProjectLog, err
	}


}

func NewMongoDbLogger(logfile string, tracelevel int) (*Logger, error) {
	templog := new(Logger)
	templog.LogFile = logfile
	templog.TraceLevel = tracelevel
	if w, err := templog.getWriter(); err != nil {
		return templog, err
	} else {
		templog.trace = log.New(w, "[Mongodb] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
		return templog, err
	}
}

func (l *Logger) Traceln(v ...interface{}) {
	l.outputln(l.trace, l.TraceLevel, v...)
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	l.outputf(l.trace, l.TraceLevel, format, v...)
}

func (l *Logger) Infoln(v ...interface{}) {
	l.outputln(l.info, l.TraceLevel, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.outputf(l.info, l.TraceLevel, format, v...)
}

func (l *Logger) Warnln(v ...interface{}) {
	l.outputln(l.warn, l.TraceLevel, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.outputf(l.warn, l.TraceLevel, format, v...)
}

func (l *Logger) Errorln(v ...interface{}) {
	l.outputln(l.error, l.TraceLevel, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.outputf(l.error, l.TraceLevel, format, v...)
}

func (l *Logger) outputln(logger *log.Logger, tracelevel int, v ...interface{}) {
	s := fmt.Sprintln(v...) + l.getTraceInfo(tracelevel)
	logger.Output(3, s)
}

func (l *Logger) outputf(logger *log.Logger, tracelevel int, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...) + l.getTraceInfo(tracelevel)
	logger.Output(3, s)
}

func (l *Logger) getWriter() (io.Writer, error) {
	lf := l.LogFile
	if lf == "" {
		return os.Stdout, nil
	}
	return os.OpenFile(l.LogFile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

func (l *Logger) getTraceInfo(level int) string {
	t := ""
	return t
}

//mongoDB
func (l *Logger) Output(calldepth int, s string) error {
	//
	//for i := 0; i < calldepth; i++ {
	//	_, _, _, ok := runtime.Caller(3 + i)
	//	if !ok {
	//		break
	//	}
	//
	//}
	l.trace.Output(calldepth+1, s)
	return nil
}
