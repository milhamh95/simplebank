package gapi

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	userAgentHeader            = "user-agent"
)

func (s *Server) extractMetadata(ctx context.Context) (*Metadata, error) {
	loginMetadata := Metadata{}

	reqMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("metadata is not found")
	}

	grpcUserAgent := reqMetadata.Get(grpcGatewayUserAgentHeader)
	if len(grpcUserAgent) > 0 {
		loginMetadata.UserAgent = grpcUserAgent[0]
	}

	httpUserAgent := reqMetadata.Get(userAgentHeader)
	if len(userAgentHeader) > 0 {
		loginMetadata.UserAgent = httpUserAgent[0]
	}

	clientIP := reqMetadata.Get(xForwardedForHeader)
	if len(clientIP) > 0 {
		loginMetadata.ClientIP = clientIP[0]
	}

	// peer info  contains the information of the peer for an RPC, such as the address
	// and authentication information.
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("rpc peer info is not found")
	}
	loginMetadata.ClientIP = peerInfo.Addr.String()

	return &loginMetadata, nil
}
