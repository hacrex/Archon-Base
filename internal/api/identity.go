package api

import (
	"regexp"
	"strings"
	"time"
)

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Membership struct {
	UserID       string    `json:"userId"`
	Organization string    `json:"organizationId"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ProjectMembership struct {
	Project     string    `json:"project"`
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (u User) Validate() error {
	var errs ValidationErrors
	if !emailPattern.MatchString(strings.ToLower(strings.TrimSpace(u.Email))) {
		errs = errs.Add("email", "must be a valid email address")
	}
	if strings.TrimSpace(u.DisplayName) == "" {
		errs = errs.Add("displayName", "is required")
	}
	if u.Status != "" && u.Status != "active" && u.Status != "invited" && u.Status != "suspended" {
		errs = errs.Add("status", "must be active, invited, or suspended")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (m Membership) Validate() error {
	var errs ValidationErrors
	if strings.TrimSpace(m.UserID) == "" {
		errs = errs.Add("userId", "is required")
	}
	if strings.TrimSpace(m.Organization) == "" {
		errs = errs.Add("organizationId", "is required")
	}
	switch m.Role {
	case "owner", "admin", "operator", "developer", "viewer":
	default:
		errs = errs.Add("role", "must be owner, admin, operator, developer, or viewer")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (m ProjectMembership) Validate() error {
	var errs ValidationErrors
	if strings.TrimSpace(m.UserID) == "" {
		errs = errs.Add("userId", "is required")
	}
	switch m.Role {
	case "owner", "admin", "operator", "developer", "viewer":
	default:
		errs = errs.Add("role", "must be owner, admin, operator, developer, or viewer")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
