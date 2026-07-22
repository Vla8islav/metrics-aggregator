package middlewares_grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func WithLogging(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		st := status.Convert(err)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("grpc_code", st.Code().String()),
			zap.Duration("duration", time.Since(start)),
		}

		if err != nil {
			logger.Warn(
				"gRPC request completed with error",
				append(fields, zap.Error(err))...,
			)
		} else {
			logger.Info("gRPC request completed", fields...)
		}

		return resp, err
	}
}
