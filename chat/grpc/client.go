package grpc

import (
	"context"
	"time"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Client(token string) (string, error) {
	addr := os.Getenv("AUTH_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := client.GetUserId(ctx, &TokenRequest{Token: token})
	if err != nil {
		return "", err
	}

	return res.UserId, nil
}