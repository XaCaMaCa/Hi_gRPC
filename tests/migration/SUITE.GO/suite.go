package suite

import (
	"context"
	"fmt"
	"sso/internal/config"
	"testing"

	authv1 "github.com/XaCaMaCa/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Config     *config.Config
	AuthClient authv1.AuthServiceClient
}

func New(t *testing.T) (context.Context, *Suite) {

	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.Grpc.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	conn, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Config:     cfg,
		AuthClient: authv1.NewAuthServiceClient(conn),
	}
}

func grpcAddress(cfg *config.Config) string {
	return fmt.Sprintf("localhost:%d", cfg.Grpc.Port)
}
