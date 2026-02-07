package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

type PasswordManager struct {
	cost int
}

func NewPasswordManager(cost int) *PasswordManager {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &PasswordManager{cost: cost}
}

func (pm *PasswordManager) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (pm *PasswordManager) Compare(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return ErrInvalidPassword
	}
	return nil
}
