package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func initSystem() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file")
	}
	return nil
}

func initLogger() *zap.Logger {
	// Создаем папку logs, если ее нет
	_ = os.MkdirAll("logs", 0755)

	// Основной файл: ротация каждые 3 часа
	mainFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxAge:     7,  // хранить 7 дней
		MaxBackups: 56, // 24/3 * 7 = 56 ротаций
		Compress:   false,
	}

	// Файл ошибок
	errorFile := &lumberjack.Logger{
		Filename:   "logs/errors.log",
		MaxAge:     7,
		MaxBackups: 56,
		Compress:   false,
	}

	// Настройка формата JSON
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// Потоки
	consoleWS := zapcore.AddSync(os.Stdout)
	mainWS := zapcore.AddSync(mainFile)
	errorWS := zapcore.AddSync(errorFile)

	// Уровни
	infoLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.InfoLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	})

	// Основной поток: INFO+ в файл и в консоль
	mainCore := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleWS, infoLevel),
		zapcore.NewCore(encoder, mainWS, infoLevel),
	)

	// Дополнительный поток ошибок: ERROR+ в отдельный файл
	errorCore := zapcore.NewCore(encoder, errorWS, errorLevel)

	// Объединяем
	core := zapcore.NewTee(mainCore, errorCore)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

func main() {
	logger := initLogger()
	defer logger.Sync()

	for i := range 3 {
		fmt.Println(i)
	}

}
