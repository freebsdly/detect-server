package log

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

var Logger *zap.SugaredLogger

func InitLogger() {
	var (
		logLevel = viper.GetString("log.level")
		logPath  = viper.GetString("log.path")
	)
	//获取编码器,NewJSONEncoder()输出json格式，NewConsoleEncoder()输出普通文本格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder //指定时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//文件writeSyncer
	fileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath, //日志文件存放目录
		MaxSize:    10,      //文件大小限制,单位MB
		MaxBackups: 10,      //最大保留日志文件数量
		MaxAge:     30,      //日志文件保留天数
		Compress:   false,   //是否压缩处理
	})
	var zapLogLevel = zapcore.WarnLevel
	if strings.ToLower(logLevel) == "info" {
		zapLogLevel = zapcore.InfoLevel
	} else if strings.ToLower(logLevel) == "debug" {
		zapLogLevel = zapcore.DebugLevel
	} else {
		zapLogLevel = zapcore.DebugLevel
	}
	switch strings.ToLower(logLevel) {
	case "info", "information":
		zapLogLevel = zapcore.InfoLevel
		break
	case "warn", "warning":
		zapLogLevel = zapcore.WarnLevel
	case "error":
		zapLogLevel = zapcore.ErrorLevel
	case "panic", "critical":
		zapLogLevel = zapcore.PanicLevel
	default:
		zapLogLevel = zapcore.DebugLevel
	}
	fileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(fileWriteSyncer, zapcore.AddSync(os.Stdout)), zapLogLevel) //第三个及之后的参数为写入文件的日志级别,ErrorLevel模式只记录error级别的日志

	Logger = zap.New(fileCore, zap.AddCaller()).Sugar() //AddCaller()为显示文件名和行号
}
