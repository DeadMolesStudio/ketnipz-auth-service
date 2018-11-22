package main

import (
	"net"

	"google.golang.org/grpc"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"
)

func main() {
	l := logger.InitLogger()
	defer l.Sync()

	sm := NewSessionManager("user@redis:6379", "0")
	defer sm.Close()

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Panicf("cant listen port %v", err)
	}

	server := grpc.NewServer()

	session.RegisterSessionManagerServer(server, sm)

	logger.Info("starting server at: ", 8081)
	logger.Panic(server.Serve(lis))
}
