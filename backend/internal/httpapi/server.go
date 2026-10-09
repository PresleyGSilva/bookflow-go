package httpapi

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Pool      *pgxpool.Pool
	JWTSecret string
}

func New(pool *pgxpool.Pool, secret string) *Server {
	return &Server{Pool: pool, JWTSecret: secret}
}
