package grpc

import (
	"context"
	"net"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryCIDRInterceptor(allowedCIDR string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var realIP string
		if md, exists := metadata.FromIncomingContext(ctx); exists {
			ip := md.Get("x-real-ip")
			if len(ip) > 0 {
				realIP = ip[0]
			}
		}
		if len(realIP) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}
		netIP, err := utils.ParseCIDRString(allowedCIDR)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid allowed CIDR configuration")
		}
		if !utils.IsIPInCIDR(net.ParseIP(realIP), netIP) {
			return nil, status.Error(codes.PermissionDenied, "access from your IP is denied")
		}

		return handler(ctx, req)
	}
}
