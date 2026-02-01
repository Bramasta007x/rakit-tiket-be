package util

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	publicConstant "rakit-tiket-be/pkg/constant"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFunc func(ctx context.Context) []zapcore.Field

type LogUtil struct {
	*zap.Logger
	logFunc []LogFunc
	Level   string
}

func MakeLogUtil(level string, logFile *string, options ...zap.Option) LogUtil {
	// Define log level
	logLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// Configure console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	if strings.EqualFold(level, publicConstant.LogDevelopment) {
		// Define log level
		logLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		// Configure console encoder
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	// Create console output sink
	consoleOutput := zapcore.Lock(os.Stdout)
	zapCores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, consoleOutput, logLevel),
	}
	if logFile != nil {
		// Create a lumberjack logger for log rotation
		lumberjackLogger := &lumberjack.Logger{
			Filename:   *logFile, // Log file path
			MaxSize:    10,       // Max size before rotation (in megabytes)
			MaxBackups: 5,        // Max number of backups
			MaxAge:     30,       // Max days to retain old logs
			Compress:   true,     // Compress backups
		}
		fileOutput := zapcore.AddSync(lumberjackLogger)
		// // Configure file encoder
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		// // Create file output sink
		// // logFile, _ := os.Create(filePath)
		// // defer logFile.Close()
		// fileOutput := zapcore.AddSync(logFile)
		// // Create core with multiple outputs
		zapCores = append(zapCores, zapcore.NewCore(fileEncoder, fileOutput, logLevel))
	}

	// Create logger
	cores := zapcore.NewTee(zapCores...)
	zapLog := zap.New(cores, options...)
	return LogUtil{
		Logger: zapLog,
		Level:  level,
	}
}

func NewLogFile(filePath, fileName string) *os.File {
	savePath := fmt.Sprintf("%s/%s.log", filePath, fileName)
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		logFile, err := os.Create(savePath)
		if err != nil {
			return nil
		}
		return logFile
	}

	logFile, err := os.OpenFile(savePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return nil
	}
	return logFile
}

// func MakeLogUtil(env string, options ...zap.Option) LogUtil {
// 	var err error
// 	var zapLog *zap.Logger
// 	switch env {
// 	case publicConstant.LogDevelopment:
// 		cfg := zap.NewDevelopmentConfig()
// 		zapLog, err = cfg.Build(options...)
// 	default:
// 		cfg := zap.NewProductionConfig()
// 		zapLog, err = cfg.Build(options...)
// 	}

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return LogUtil{
// 		Logger: zapLog,
// 	}
// }

func (l *LogUtil) GetLevel() string {
	return l.Level
}

func (l *LogUtil) AddLogFunc(logFuncs ...LogFunc) LogUtil {
	if l.logFunc == nil {
		l.logFunc = make([]LogFunc, 0)
	}
	l.logFunc = append(l.logFunc, logFuncs...)

	return *l
}

func (l LogUtil) setExtraFields(ctx context.Context, caller string, fields ...zap.Field) []zap.Field {
	if len(l.logFunc) > 0 {
		for _, logFunc := range l.logFunc {
			fields = append(logFunc(ctx), fields...)
		}
	}
	if caller != "" {
		fields = append(fields, zap.String("caller", caller))
	}
	return fields
}

func (l LogUtil) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Info(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Error(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.DPanic(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Panic(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, l.setExtraFields(ctx, l.decodeCaller(runtime.Caller(1)), fields...)...)
}

func (l LogUtil) decodeCaller(pc uintptr, file string, line int, ok bool) string {
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}
