package service

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	"onlinejudge/internal/model"
	"onlinejudge/pkg/idgen"
)

// HashPassword 使用 SHA-256 生成密码摘要（演示用途，非生产级）。
func HashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (s *Service) CreateUser(u model.User, password string) (*model.User, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetUserByUsername(u.Username); err == nil {
		return nil, model.NewValidationError("username", "用户名已存在")
	}
	if password == "" {
		return nil, model.NewValidationError("password", "密码不能为空")
	}
	u.ID = idgen.Hex()
	u.PasswordHash = HashPassword(password)
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	if err := s.store.CreateUser(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) GetUser(id string) (*model.User, error) {
	return s.store.GetUser(id)
}

func (s *Service) ListUsers(filter model.UserFilter, page, size int) ([]*model.User, int, error) {
	all := s.store.ListUsers()
	matched := make([]*model.User, 0, len(all))
	for _, u := range all {
		if filter.Match(u) {
			matched = append(matched, u)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.Before(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.User{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateUser(id string, u model.User) (*model.User, error) {
	existing, err := s.store.GetUser(id)
	if err != nil {
		return nil, err
	}
	if u.Email != "" {
		existing.Email = u.Email
	}
	if u.Role != "" {
		existing.Role = u.Role
	}
	if u.Status != "" {
		existing.Status = u.Status
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateUser(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteUser(id string) error {
	return s.store.DeleteUser(id)
}

// Authenticate 校验用户名密码，返回用户。
func (s *Service) Authenticate(username, password string) (*model.User, error) {
	u, err := s.store.GetUserByUsername(username)
	if err != nil {
		return nil, model.NewValidationError("username", "用户名或密码错误")
	}
	if u.PasswordHash != HashPassword(password) {
		return nil, model.NewValidationError("password", "用户名或密码错误")
	}
	if u.Status != model.UserActive {
		return nil, model.NewValidationError("status", "账号已被禁用")
	}
	return u, nil
}

// ChangePassword 修改用户密码，需校验旧密码。
func (s *Service) ChangePassword(id, oldPassword, newPassword string) error {
	u, err := s.store.GetUser(id)
	if err != nil {
		return err
	}
	if u.PasswordHash != HashPassword(oldPassword) {
		return model.NewValidationError("old_password", "旧密码错误")
	}
	if newPassword == "" {
		return model.NewValidationError("new_password", "新密码不能为空")
	}
	u.PasswordHash = HashPassword(newPassword)
	u.UpdatedAt = time.Now()
	return s.store.UpdateUser(u)
}
