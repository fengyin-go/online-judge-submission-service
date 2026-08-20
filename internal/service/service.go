package service

import (
	"onlinejudge/internal/config"
	"onlinejudge/internal/store"
	"onlinejudge/pkg/logger"
)

type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}

// BatchFailure 批量操作中单个条目的失败信息。
type BatchFailure struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}
