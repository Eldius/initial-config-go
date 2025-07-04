package telemetry

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newGrpcConnection(endpoint string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
