package api

import "testing"

func TestUserValidation(t *testing.T) {
	if err := (User{Email: "student@example.com", DisplayName: "Student", Status: "invited"}).Validate(); err != nil {
		t.Fatalf("valid user rejected: %v", err)
	}
	if err := (User{Email: "not-an-email", DisplayName: ""}).Validate(); err == nil {
		t.Fatal("invalid user accepted")
	}
}

func TestMembershipValidation(t *testing.T) {
	if err := (Membership{UserID: "user-1", Organization: "org-1", Role: "developer"}).Validate(); err != nil {
		t.Fatalf("valid membership rejected: %v", err)
	}
	if err := (Membership{UserID: "user-1", Organization: "org-1", Role: "superuser"}).Validate(); err == nil {
		t.Fatal("invalid membership role accepted")
	}
}
