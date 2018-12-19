package main

import (
	"flag"
	"net"

	"google.golang.org/grpc"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"
)

func main() {
	dbConnStr := flag.String("db_connstr", "user@localhost:6379", "redis connection string")
	dbName := flag.String("db_name", "0", "redis database name")
	flag.Parse()

	l := logger.InitLogger()
	defer func() {
		err := l.Sync()
		if err != nil {
			logger.Errorf("error while syncing log data: %v", err)
		}
	}()

	sm := NewSessionManager(*dbConnStr, *dbName)
	defer sm.Close()

	/* #nosec */
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Panicf("cant listen port %v", err)
	}

	server := grpc.NewServer()

	session.RegisterSessionManagerServer(server, sm)

	logger.Info("starting server at: ", 8081)
	logger.Panic(server.Serve(lis))
}
