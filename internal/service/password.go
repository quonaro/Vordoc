// Package service implements application business logic for Vordoc.
package service

import "golang.org/x/crypto/bcrypt"

// PasswordService verifies bcrypt password hashes.
type PasswordService struct{}

// NewPasswordService creates a password service.
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// Hash creates a bcrypt hash from a plaintext password.
func (s *PasswordService) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify compares a plaintext password against a bcrypt hash.
func (s *PasswordService) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
