package session

import (
	"context"
	"time"

	grpc "google.golang.org/grpc"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
)

var (
	sm SessionManagerClient
)

func ConnectSessionManager() *grpc.ClientConn {
	grpcConn, err := grpc.Dial(
		"auth-service:8081",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(30*time.Second),
	)
	if err != nil {
		logger.Panic("failed to connect to sessionManager: ", err)
	}

	sm = NewSessionManagerClient(grpcConn)

	logger.Infof("Successfully connected to sessionManager: %v", 8081)

	return grpcConn
}

func Create(uID uint) (string, error) {
	sID, err := sm.Create(
		context.Background(),
		&Session{UID: uint64(uID)},
	)
	if err != nil {
		return "", err
	}
	return sID.UUID, nil
}

func Get(sID string) (uint, error) {
	s, err := sm.Get(
		context.Background(),
		&SessionID{UUID: sID},
	)
	if err != nil {
		return 0, err
	}
	return uint(s.UID), nil
}

func Delete(sID string) error {
	_, err := sm.Delete(
		context.Background(),
		&SessionID{UUID: sID},
	)
	return err
}
