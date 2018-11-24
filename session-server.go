package main

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"
)

type SessionManager struct {
	redisConnPool *redis.Pool
}

func NewSessionManager(address, database string) *SessionManager {
	sm := &SessionManager{}
	err := sm.Open(address, database)
	if err != nil {
		logger.Panic(err)
	}

	logger.Infof("Successfully connected to %v, database %v", address, database)

	return sm
}

func (sm *SessionManager) Open(address, database string) error {
	sm.redisConnPool = &redis.Pool{
		MaxIdle: 500,
		IdleTimeout: 240 * time.Second,
		MaxActive: 1000,
		Wait: true,
		Dial: func () (redis.Conn, error) { return redis.DialURL("redis://" + address + "/" + database) },
	}
	conn := sm.redisConnPool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	return err
}

func (sm *SessionManager) Close() {
	sm.redisConnPool.Close()
}

func (sm *SessionManager) Create(ctx context.Context, in *session.Session) (*session.SessionID, error) {
	sID := ""
	conn := sm.redisConnPool.Get()
	defer conn.Close()
	for {
		sID = createUUID()
		res, err := conn.Do("SET", sID, in.UID, "NX", "EX", 30*24*60*60)
		if err != nil {
			return &session.SessionID{}, status.Error(codes.Internal, err.Error())
		}
		if res != "OK" {
			logger.Infow("collision, session not created",
				"sID", sID,
				"uID", in.UID,
			)
			continue
		}
		break
	}

	logger.Infow("session created",
		"sID", sID,
		"uID", in.UID,
	)

	return &session.SessionID{UUID: sID}, nil
}

func (sm *SessionManager) Get(ctx context.Context, in *session.SessionID) (*session.Session, error) {
	conn := sm.redisConnPool.Get()
	defer conn.Close()
	res, err := redis.Uint64(conn.Do("GET", in.UUID))
	if err != nil {
		if err == redis.ErrNil {
			return &session.Session{}, status.Error(codes.NotFound, session.ErrKeyNotFound.Error())
		}
		return &session.Session{}, status.Error(codes.Internal, err.Error())
	}

	return &session.Session{UID: res}, nil
}

func (sm *SessionManager) Delete(ctx context.Context, in *session.SessionID) (*session.Nothing, error) {
	conn := sm.redisConnPool.Get()
	defer conn.Close()
	_, err := redis.Int(conn.Do("DEL", in.UUID))
	if err != nil {
		return &session.Nothing{}, status.Error(codes.Internal, err.Error())
	}

	logger.Infow("session deleted",
		"sID", in.UUID,
	)

	return &session.Nothing{}, nil
}

func createUUID() string {
	return uuid.NewV4().String()
}
