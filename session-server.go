package main

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"
)

type SessionManager struct {
	redisConn redis.Conn
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
	var err error
	sm.redisConn, err = redis.DialURL("redis://" + address + "/" + database)
	return err
}

func (sm *SessionManager) Close() {
	sm.redisConn.Close()
}

func (sm *SessionManager) Create(ctx context.Context, in *session.Session) (*session.SessionID, error) {
	sID := ""
	for {
		sID = createUUID()
		res, err := sm.redisConn.Do("SET", sID, in.UID, "NX", "EX", 30*24*60*60)
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
	res, err := redis.Uint64(sm.redisConn.Do("GET", in.UUID))
	if err != nil {
		if err == redis.ErrNil {
			return &session.Session{}, status.Error(codes.NotFound, session.ErrKeyNotFound.Error())
		}
		return &session.Session{}, status.Error(codes.Internal, err.Error())
	}

	return &session.Session{UID: res}, nil
}

func (sm *SessionManager) Delete(ctx context.Context, in *session.SessionID) (*session.Nothing, error) {
	_, err := redis.Int(sm.redisConn.Do("DEL", in.UUID))
	if err != nil {
		return &session.Nothing{}, status.Error(codes.Internal, err.Error())
	}

	return &session.Nothing{}, nil
}

func createUUID() string {
	return uuid.NewV4().String()
}
