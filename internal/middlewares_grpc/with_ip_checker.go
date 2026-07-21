package middlewares_grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// WithIPChecker creates a unary gRPC interceptor that allows requests only
// from addresses belonging to trustedSubnet, for example 192.168.1.0/24.
func WithIPChecker(
	subnet config.OptionalString,
	logger *zap.Logger,
) (grpc.UnaryServerInterceptor, error) {
	if !subnet.BeenSet {
		return nil, fmt.Errorf(
			"trusted subnet is not set",
		)
	}

	_, network, err := net.ParseCIDR(subnet.Value)
	if err != nil {
		return nil, fmt.Errorf(
			"parse trusted subnet %q: %w",
			subnet,
			err,
		)
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		clientIP, err := clientIPFromContext(ctx)
		if err != nil {
			logger.Warn(
				"failed to determine gRPC client IP",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)

			return nil, status.Error(
				codes.PermissionDenied,
				"unable to determine client address",
			)
		}

		if !network.Contains(clientIP) {
			logger.Warn(
				"gRPC request rejected: IP is outside trusted subnet",
				zap.String("method", info.FullMethod),
				zap.String("client_ip", clientIP.String()),
				zap.String("trusted_subnet", network.String()),
			)

			return nil, status.Error(
				codes.PermissionDenied,
				"client IP is outside the trusted subnet",
			)
		}

		return handler(ctx, req)
	}, nil
}

func clientIPFromContext(ctx context.Context) (net.IP, error) {
	remotePeer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("peer information is missing")
	}

	if remotePeer.Addr == nil {
		return nil, fmt.Errorf("peer address is missing")
	}

	if tcpAddr, ok := remotePeer.Addr.(*net.TCPAddr); ok {
		if tcpAddr.IP == nil {
			return nil, fmt.Errorf("TCP peer IP is missing")
		}

		return tcpAddr.IP, nil
	}

	host, _, err := net.SplitHostPort(remotePeer.Addr.String())
	if err != nil {
		return nil, fmt.Errorf(
			"parse peer address %q: %w",
			remotePeer.Addr.String(),
			err,
		)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid peer IP %q", host)
	}

	return ip, nil
}
