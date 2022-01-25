package slog

import (
	"fmt"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type LogOption struct {
	Console      bool   `yaml:"console" json:"console"`
	File         bool   `yaml:"file" json:"file"`
	Level        string `yaml:"level" json:"level"`
	Path         string `yaml:"path" json:"path"`
	RollbackTime int    `yaml:"rollback_time" json:"rollback_time"`
	Count        int    `yaml:"count" json:"count"`
}

// 自定义日志格式
type myFormatter struct{}

// type logMessage struct {
// 	AppName   interface{} `json:"appName"`
// 	Timestamp interface{} `json:"timestamp"`
// 	Level     interface{} `json:"level"`
// 	Message   interface{} `json:"message"`
// 	Type      interface{} `json:"Type"`
// 	TraceId   interface{} `json:"traceId"`
// 	SpanId    interface{} `json:"spanId"`
// 	ParentId  interface{} `json:"parentId"`
// 	Host      interface{} `json:"host"`
// 	Ip        interface{} `json:"ip"`
// }

func (s myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := fmt.Sprintf("%v", entry.Level)
	msg := fmt.Sprintf("%-24s %-9s %-8s %s\n", time.Now().Format("2006-01-02T15:04:05.999"), strings.ToUpper(level), entry.Data["type"], entry.Message)
	return []byte(msg), nil
}

func setLogLevel(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warm":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "trace":
		return logrus.TraceLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		logrus.Warn("日志级别设置错误，使用默认日志级别:\"info\"")
		return logrus.InfoLevel
	}
}

func NewLogger(logOption *LogOption) (*logrus.Logger, error) {
	//var err error
	var log = logrus.New()
	log.SetReportCaller(true)
	// 设置日志级别为xx以及以上
	log.SetLevel(setLogLevel(logOption.Level))
	//log.AddHook(&defaultFieldHook{})
	// 设置日志格式为json格式
	//log.SetFormatter(&logrus.JSONFormatter{
	//	// PrettyPrint: true,//格式化json
	//	TimestampFormat: "2006-01-02 15:04:05",//时间格式化
	//})
	//log.SetFormatter(&logrus.TextFormatter{
	//	ForceColors:               true,
	//	EnvironmentOverrideColors: true,
	//	// FullTimestamp:true,
	//	TimestampFormat: "2006-01-02 15:04:05", //时间格式化
	//	// DisableLevelTruncation:true,
	//})
	log.SetFormatter(myFormatter{})
	// 设置将日志输出到标准输出（默认的输出为stdout，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	// file, _ := os.OpenFile("/tmp/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if logOption.Console {
		log.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		log.SetOutput(file)
	}
	if logOption.Path != "" {
		writer, err := rotatelogs.New(
			//这是分割代码的命名规则，要和下面WithRotationTime时间精度一致
			logOption.Path+".%Y%m%d%H%M%S",
			// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件。
			rotatelogs.WithLinkName(logOption.Path),
			//文件切割之间的间隔。默认情况下，日志每86400秒/一天旋转一次。注意:记住要利用时间。持续时间值。
			rotatelogs.WithRotationTime(time.Duration(logOption.RollbackTime)*time.Second*86400),
			// WithMaxAge和WithRotationCount二者只能设置一个，
			// WithMaxAge设置文件清理前的最长保存时间，
			// WithRotationCount设置文件清理前最多保存的个数。 默认情况下，此选项是禁用的。
			// rotatelogs.WithMaxAge(time.Second*30), //默认每7天清除下日志文件
			rotatelogs.WithRotationCount(uint(logOption.Count)),
			//rotatelogs.WithMaxAge(-1),       //需要手动禁用禁用  默认情况下不清除日志，
			// rotatelogs.WithRotationCount(2), //清除除最新2个文件之外的日志，默认禁用
		)
		if err != nil {
			return nil, err
		}

		lfsHook := lfshook.NewHook(lfshook.WriterMap{
			logrus.TraceLevel: writer,
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}, &myFormatter{})
		log.AddHook(lfsHook)
	}
	return log, nil
}
