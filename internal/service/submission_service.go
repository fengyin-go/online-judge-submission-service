package service

import (
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

func (s *Service) CreateSubmission(sub model.Submission) (*model.Submission, error) {
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetProblem(sub.ProblemID); err != nil {
		return nil, model.NewValidationError("problem_id", "题目不存在")
	}
	if _, err := s.store.GetUser(sub.UserID); err != nil {
		return nil, model.NewValidationError("user_id", "用户不存在")
	}
	sub.ID = idgen.Hex()
	sub.Status = model.SubmissionPending
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = sub.CreatedAt
	if err := s.store.CreateSubmission(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *Service) GetSubmission(id string) (*model.Submission, error) {
	return s.store.GetSubmission(id)
}

// GetJudgeResult 返回某提交的判题明细。
func (s *Service) GetJudgeResult(submissionID string) (*model.JudgeResult, error) {
	return s.store.GetJudgeResultBySubmission(submissionID)
}

func (s *Service) ListSubmissions(filter model.SubmissionFilter, page, size int) ([]*model.Submission, int, error) {
	all := s.store.ListSubmissions()
	matched := make([]*model.Submission, 0, len(all))
	for _, sub := range all {
		if filter.Match(sub) {
			matched = append(matched, sub)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Submission{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// JudgeSubmission 执行判题：pending -> judging -> 终态，并写入判题明细。
func (s *Service) JudgeSubmission(id string) (*model.JudgeResult, error) {
	sub, err := s.store.GetSubmission(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionSubmission(sub.Status, model.SubmissionJudging) {
		return nil, model.NewValidationError("status", "当前状态不可判题")
	}
	sub.Status = model.SubmissionJudging
	sub.UpdatedAt = time.Now()
	_ = s.store.UpdateSubmission(sub)

	// 模拟判题耗时。
	if s.cfg != nil && s.cfg.JudgeDelayMs > 0 {
		time.Sleep(time.Duration(s.cfg.JudgeDelayMs) * time.Millisecond)
	}

	timeLimit := 1000
	if p, err := s.store.GetProblem(sub.ProblemID); err == nil && p.TimeLimitMs > 0 {
		timeLimit = p.TimeLimitMs
	}
	verdict, timeUsed, memUsed := simulateJudge(sub, timeLimit)
	sub.Status = verdict
	sub.TimeUsedMs = timeUsed
	sub.MemoryUsedKB = memUsed
	sub.UpdatedAt = time.Now()
	if err := s.store.UpdateSubmission(sub); err != nil {
		return nil, err
	}

	j := &model.JudgeResult{
		ID:           idgen.Hex(),
		SubmissionID: sub.ID,
		Verdict:      verdict,
		TimeUsedMs:   timeUsed,
		MemoryUsedKB: memUsed,
		Message:      verdictMessage(verdict),
		JudgedAt:     time.Now(),
		CreatedAt:    time.Now(),
	}
	if err := s.store.CreateJudgeResult(j); err != nil {
		return nil, err
	}
	return j, nil
}

// simulateJudge 根据代码特征确定性模拟判题结果，便于演示与测试。
func simulateJudge(sub *model.Submission, timeLimit int) (verdict string, timeUsed, memUsed int) {
	sum := 0
	for _, b := range []byte(sub.Code) {
		sum += int(b)
	}
	timeUsed = 1 + sum%timeLimit
	if timeUsed == 0 {
		timeUsed = 1
	}
	memUsed = 1024 + sum%subMemoryGuess(sub)
	switch sum % 6 {
	case 0:
		return model.SubmissionAccepted, timeUsed, memUsed
	case 1:
		return model.SubmissionWrongAnswer, timeUsed, memUsed
	case 2:
		return model.SubmissionTimeLimit, timeLimit, memUsed
	case 3:
		return model.SubmissionRuntimeError, timeUsed, memUsed
	case 4:
		return model.SubmissionCompileError, 0, 0
	default:
		return model.SubmissionMemoryLimit, timeUsed, memUsed
	}
}

func subMemoryGuess(sub *model.Submission) int {
	if sub.MemoryUsedKB > 0 {
		return sub.MemoryUsedKB
	}
	return 65536
}

func verdictMessage(v string) string {
	switch v {
	case model.SubmissionAccepted:
		return "通过全部测试用例"
	case model.SubmissionWrongAnswer:
		return "输出与预期不符"
	case model.SubmissionTimeLimit:
		return "运行时间超出限制"
	case model.SubmissionMemoryLimit:
		return "内存使用超出限制"
	case model.SubmissionRuntimeError:
		return "运行时错误"
	case model.SubmissionCompileError:
		return "编译失败"
	default:
		return ""
	}
}

func (s *Service) DeleteSubmission(id string) error {
	return s.store.DeleteSubmission(id)
}

// RejudgeSubmission 对终态提交重新判题：重置为 pending 后重新走判题流程。
func (s *Service) RejudgeSubmission(id string) (*model.JudgeResult, error) {
	sub, err := s.store.GetSubmission(id)
	if err != nil {
		return nil, err
	}
	if !model.IsFinalVerdict(sub.Status) {
		return nil, model.NewValidationError("status", "仅终态提交可重判")
	}
	sub.Status = model.SubmissionPending
	sub.UpdatedAt = time.Now()
	if err := s.store.UpdateSubmission(sub); err != nil {
		return nil, err
	}
	if old, err := s.store.GetJudgeResultBySubmission(sub.ID); err == nil {
		_ = s.store.DeleteJudgeResult(old.ID)
	}
	return s.JudgeSubmission(sub.ID)
}
