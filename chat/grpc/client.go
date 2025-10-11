package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Client(token string) (string, error) {
	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

	res, err := client.GetUserId(ctx, &TokenRequest{Token: token})
	if err != nil {
		return "", err
	}

	return res.UserId, nil
}