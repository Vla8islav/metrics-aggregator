// Package interceptors_grpc ip checker grpc interceptor
package interceptors_grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// WithIPChecker creates a unary gRPC interceptor that allows requests only
// from addresses belonging to trustedSubnet, for example 192.168.1.0/24.
func WithIPChecker(
	subnets config.OptionalString,
	logger *zap.Logger,
) (grpc.UnaryServerInterceptor, error) {
	if !subnets.BeenSet {
		return nil, fmt.Errorf(
			"trusted subnets is not set",
		)
	}

	if strings.TrimSpace(subnets.Value) == "" {
		return nil, fmt.Errorf("trusted subnets is empty")
	}

	var networks []*net.IPNet

	for _, rawSubnet := range strings.Split(subnets.Value, ",") {
		subnet := strings.TrimSpace(rawSubnet)
		_, network, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, fmt.Errorf(
				"parse trusted subnet faulted %q: %w",
				subnet,
				err,
			)
		}
		networks = append(networks, network)
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		clientIP, err := extractRealIpFromMetadata(ctx)
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

		for _, network := range networks {
			if network.Contains(clientIP) {
				return handler(ctx, req)
			}
		}

		logger.Warn(
			"gRPC request rejected: IP is outside trusted subnets",
			zap.String("method", info.FullMethod),
			zap.String("client_ip", clientIP.String()),
			zap.String("trusted_subnets", subnets.Value),
		)

		return nil, status.Error(
			codes.PermissionDenied,
			"client IP is outside the trusted subnets",
		)

	}, nil
}

func extractRealIpFromMetadata(ctx context.Context) (net.IP, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is missing")
	}
	values := md.Get("x-real-ip")
	if len(values) > 0 {
		clientIpStr := values[0]

		clientIP := net.ParseIP(clientIpStr)
		if clientIP == nil {
			return nil, fmt.Errorf("invalid x-real-ip header value %q", clientIpStr)
		}

		return clientIP, nil
	}
	return nil, fmt.Errorf("x-real-ip header is missing")
}
