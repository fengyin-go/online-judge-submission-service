package service

import (
	"context"
	"errors"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"reflect"
)

const flowServiceVariant = "broken-010"

type FlowService struct {
	store     *store.FlowStore
	validator model.FlowValidator
}

func NewFlowService(st *store.FlowStore, v model.FlowValidator) *FlowService {
	if v == nil {
		v = model.DefaultFlowValidator{}
	}
	if reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil() {
		v = model.RejectingFlowValidator{}
	}
	return &FlowService{store: st, validator: v}
}

func (s *FlowService) Process(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// 校验基于快照进行：被校验的字段（UserID/ProblemID）不会变化，
	// 这样无效工单不会留下中间状态。
	snap := s.store.Get(id)
	if snap == nil {
		return errors.New("ticket missing")
	}
	if err := s.validator.Validate(snap); err != nil {
		return err
	}
	// 原子地认领工单进入 judging。若工单已被并发回调认领、或已处于终态，
	// 认领失败，本次重复回调不得再产生任何结果（不增加次数、不发事件）。
	if _, ok := s.store.ClaimForJudging(id); !ok {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// 原子地提交终态并记录完成事件。若并发调用已先行提交，则本次为空操作，
	// 确保一次终态只产生一次结果。
	if !s.store.CommitTerminal(id, model.FlowAccepted, "accepted:"+id) {
		return nil
	}
	return nil
}

func (s *FlowService) Retry(ctx context.Context, id string) error {
	if t := s.store.Get(id); t != nil && t.Status == model.FlowAccepted {
		s.store.AddEvent("retry:" + id)
		return nil
	}
	if err := s.Process(ctx, id); err != nil {
		return err
	}
	s.store.AddEvent("retry:" + id)
	return nil
}

func (s *FlowService) Snapshot() []*model.FlowTicket { return s.store.List() }
