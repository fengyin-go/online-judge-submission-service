package model

import (
	"context"
	"reflect"
)

type FlowTicket struct {
	ID        string
	UserID    string
	ProblemID string
	Status    string
	Attempts  int
	Items     []string
}

const (
	FlowPending  = "pending"
	FlowJudging  = "judging"
	FlowAccepted = "accepted"
	FlowFailed   = "failed"
)

type FlowValidator interface{ Validate(*FlowTicket) error }
type DefaultFlowValidator struct{}

func IsNilFlowValidator(v FlowValidator) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func (DefaultFlowValidator) Validate(t *FlowTicket) error {
	if t == nil || t.UserID == "" || t.ProblemID == "" {
		return context.Canceled
	}
	return nil
}

type TypedNilFlowValidator struct{ ready bool }

func (v *TypedNilFlowValidator) Validate(*FlowTicket) error {
	if !v.ready {
		return nil
	}
	return nil
}

type RejectingFlowValidator struct{}

func (RejectingFlowValidator) Validate(*FlowTicket) error { return context.Canceled }

func CloneFlowTicket(t *FlowTicket) *FlowTicket {
	if t == nil {
		return nil
	}
	cp := *t
	cp.Items = append([]string(nil), t.Items...)
	return &cp
}
