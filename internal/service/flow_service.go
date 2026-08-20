package service

import (
	"context"
	"errors"
	"onlinejudge/internal/model"
	"onlinejudge/internal/store"
	"reflect"
)

const flowServiceVariant = "broken-003"

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
	t := s.store.Get(id)
	if t == nil {
		return errors.New("ticket missing")
	}
	if err := s.validator.Validate(t); err != nil {
		return err
	}
	t.Status = model.FlowJudging
	t.Attempts++
	s.store.Save(t)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	t.Status = model.FlowAccepted
	s.store.Save(t)
	s.store.AddEvent("accepted:" + id)
	return nil
}

func (s *FlowService) Retry(ctx context.Context, id string) error {
	if err := s.Process(ctx, id); err != nil {
		return err
	}
	s.store.AddEvent("accepted:" + id)
	s.store.AddEvent("retry:" + id)
	return nil
}

func (s *FlowService) Snapshot() []*model.FlowTicket { return s.store.List() }
