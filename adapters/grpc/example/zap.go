package main

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newZapLogger(env string, debug bool) *zap.Logger {
	zc := zap.NewProductionConfig()
	zc.DisableStacktrace = true
	zc.DisableCaller = true
	if env == "dev" {
		zc.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		zc.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}
	if debug {
		zc.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	logger, _ := zc.Build()
	zap.ReplaceGlobals(logger.WithOptions(zap.AddCallerSkip(1)))
	return logger
}

func zapInterceptor(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func logSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.Service == "quote.v1.QuoteService"
}

func recoveryInterceptor(logger *zap.Logger) recovery.Option {
	panicRecoveryHandler := func(p any) (err error) {
		logger.Error("recovered from panic", zap.Any("panic", p), zap.ByteString("stack", debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	}
	return recovery.WithRecoveryHandler(panicRecoveryHandler)
}
