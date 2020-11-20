package logger

import (
	"flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
)

var Logger *zap.SugaredLogger

var logPath = flag.String("log", "", "Log Path")

// 日志初始化
func InitLogger() {
	core := zapcore.NewCore(getEncoder(), getLogWriter(), zapcore.DebugLevel)
	Logger = zap.New(core, zap.AddCaller()).Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	var out io.Writer
	if len(*logPath) == 0 {
		out = os.Stdout
	} else {
		out, _ = os.Create(*logPath)
	}
	return zapcore.AddSync(out)
}
