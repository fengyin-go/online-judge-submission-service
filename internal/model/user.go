package model

import (
	"strings"
	"time"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	UserActive   = "active"
	UserDisabled = "disabled"
)

// User 系统用户。
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(u.Email)
	if u.Username == "" {
		return NewValidationError("username", "用户名不能为空")
	}
	if len(u.Username) < 3 {
		return NewValidationError("username", "用户名至少 3 个字符")
	}
	if u.Role == "" {
		u.Role = RoleUser
	}
	if u.Role != RoleAdmin && u.Role != RoleUser {
		return NewValidationError("role", "角色不合法")
	}
	if u.Status == "" {
		u.Status = UserActive
	}
	if u.Status != UserActive && u.Status != UserDisabled {
		return NewValidationError("status", "用户状态不合法")
	}
	return nil
}

// UserFilter 用户筛选条件。
type UserFilter struct {
	Role    string
	Status  string
	Keyword string
}

func (f UserFilter) Match(u *User) bool {
	if f.Role != "" && u.Role != f.Role {
		return false
	}
	if f.Status != "" && u.Status != f.Status {
		return false
	}
	if k := strings.ToLower(strings.TrimSpace(f.Keyword)); k != "" {
		if !strings.Contains(strings.ToLower(u.Username), k) &&
			!strings.Contains(strings.ToLower(u.Email), k) {
			return false
		}
	}
	return true
}
