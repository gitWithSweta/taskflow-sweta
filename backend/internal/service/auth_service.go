package service

import (
	"context"
	"errors"
	"time"

	"taskflow/internal/auth"
	"taskflow/internal/errs"
	"taskflow/internal/model"

	"github.com/google/uuid"
)

type AuthService struct {
	users    userRepository
	sessions sessionRepository
	secret   []byte
	tokenTTL time.Duration
}

func NewAuthService(users userRepository, sessions sessionRepository, secret []byte, tokenTTL time.Duration) *AuthService {
	return &AuthService{users: users, sessions: sessions, secret: secret, tokenTTL: tokenTTL}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (token string, user model.UserPublic, err error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	u, err := s.users.Create(ctx, name, email, hash)
	if err != nil {
		if errors.Is(err, errs.ErrEmailTaken) {
			return "", model.UserPublic{}, &errs.ValidationError{Fields: map[string]string{"email": "already registered"}}
		}
		return "", model.UserPublic{}, err
	}
	token, err = s.mintToken(ctx, u.ID, u.Email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	return token, model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, user model.UserPublic, err error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	if u == nil || !auth.CheckPassword(u.PasswordHash, password) {
		return "", model.UserPublic{}, errs.ErrInvalidCredentials
	}

	_ = s.sessions.DeleteExpired(ctx, u.ID)

	token, err = s.mintToken(ctx, u.ID, u.Email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	return token, model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (s *AuthService) Me(ctx context.Context, uid uuid.UUID) (model.UserPublic, error) {
	u, err := s.users.GetByID(ctx, uid)
	if err != nil {
		return model.UserPublic{}, err
	}
	if u == nil {
		return model.UserPublic{}, errs.ErrNotFound
	}
	return model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Delete(ctx, sessionID)
}

func (s *AuthService) ListUsers(ctx context.Context) ([]model.UserPublic, error) {
	return s.users.ListAllPublic(ctx, 500)
}

func (s *AuthService) mintToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	sessionID, err := s.sessions.Create(ctx, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return auth.SignToken(s.secret, userID, email, sessionID, expiresAt)
}
