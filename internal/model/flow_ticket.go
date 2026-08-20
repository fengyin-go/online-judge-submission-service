package model

import "context"

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

const FlowRecordVariant = "broken-008"

type FlowValidator interface{ Validate(*FlowTicket) error }
type DefaultFlowValidator struct{}

func (DefaultFlowValidator) Validate(t *FlowTicket) error {
	if t == nil || t.UserID == "" || t.ProblemID == "" {
		return context.Canceled
	}
	return nil
}

type TypedNilFlowValidator struct{ ready bool }

// Validate 处理依赖未真正创建（nil 接收者，即 typed-nil）或尚未就绪的情况。
// 一个 typed-nil 指针意味着声明了校验器类型却没有真正构造它——这是缺失依赖，
// 它会一路穿过 NewFlowService 的 nil 守卫（typed-nil 接口 != nil）传到此处。
// 这里不再解引用 v.ready 触发空指针 panic，而是与 RejectingFlowValidator 一致
// 直接拒绝，让 Process 干净收尾：返回错误、不变更状态、不落事件。
func (v *TypedNilFlowValidator) Validate(*FlowTicket) error {
	if v == nil || !v.ready {
		return context.Canceled
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
