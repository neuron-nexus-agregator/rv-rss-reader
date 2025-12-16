package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gafarov/rss-reader/internal/pkg/app"
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
	if err := initSystem(); err != nil {
		logger.Fatal("Failed to initialize system", zap.Error(err))
		os.Exit(1)
	}

	redis_host := os.Getenv("REDIS_HOST")
	if redis_host == "" {
		logger.Fatal("REDIS_HOST environment variable is not set")
		os.Exit(1)
	}

	redis_password := os.Getenv("REDIS_PASSWORD")
	if redis_password == "" {
		logger.Fatal("REDIS_PASSWORD environment variable is not set")
		os.Exit(1)
	}

	kafka_addr := os.Getenv("KAFKA_ADDR")
	if kafka_addr == "" {
		logger.Fatal("KAFKA_ADDR environment variable is not set")
		os.Exit(1)
	}

	kafka_topic := os.Getenv("KAFKA_TOPIC")
	if kafka_topic == "" {
		logger.Fatal("KAFKA_TOPIC environment variable is not set")
		os.Exit(1)
	}

	redisData := app.RedisData{
		Host:     redis_host,
		Password: redis_password,
	}

	kafkaData := app.KafkaData{
		Addr:  []string{kafka_addr},
		Topic: kafka_topic,
	}

	app, err := app.New(redisData, kafkaData, logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
		os.Exit(1)
	}

	rss_url := os.Getenv("RSS_URL")
	if rss_url == "" {
		logger.Fatal("RSS_URL environment variable is not set")
		os.Exit(1)
	}

	rss_code := os.Getenv("RSS_CODE")
	if rss_code == "" {
		logger.Fatal("RSS_CODE environment variable is not set")
		os.Exit(1)
	}

	delay := 1 * time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = app.Run(rss_url, "realtime:site", rss_code, delay, ctx)
	if err != nil {
		logger.Fatal("Failed to run application", zap.Error(err))
	}
}
