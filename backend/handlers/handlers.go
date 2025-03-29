package handlers

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/db/redis"
	"github.com/jackc/pgx/v5"
)

type GetHandler struct {
	DB *pgx.Conn
}

type SessionHandler struct {
	DB          *pgx.Conn
	RedisClient *redis.RedisClient
}

type UserHandler struct {
	DB *pgx.Conn
}
