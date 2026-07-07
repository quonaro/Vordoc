package service

import "golang.org/x/crypto/bcrypt"

// PasswordService verifies bcrypt password hashes.
type PasswordService struct{}

// NewPasswordService creates a password service.
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// Verify compares a plaintext password against a bcrypt hash.
func (s *PasswordService) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
