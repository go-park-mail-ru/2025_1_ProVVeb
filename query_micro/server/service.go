package query

import (
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type QueryServiceServerImpl struct {
	querypb.UnimplementedQueryServiceServer
	Repo *QueryRepo
}

func NewSessionService(repo *QueryRepo) *QueryServiceServerImpl {
	return &QueryServiceServerImpl{Repo: repo}
}
