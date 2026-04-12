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

type Auth struct {
	users    userRepo
	sessions sessionRepo
	secret   []byte
	tokenTTL time.Duration
}

func NewAuth(users userRepo, sessions sessionRepo, secret []byte, tokenTTL time.Duration) *Auth {
	return &Auth{users: users, sessions: sessions, secret: secret, tokenTTL: tokenTTL}
}

func (a *Auth) Register(ctx context.Context, name, email, password string) (token string, user model.UserPublic, err error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	u, err := a.users.Create(ctx, name, email, hash)
	if err != nil {
		if errors.Is(err, errs.ErrEmailTaken) {
			return "", model.UserPublic{}, &errs.ValidationError{Fields: map[string]string{"email": "already registered"}}
		}
		return "", model.UserPublic{}, err
	}
	token, err = a.mintToken(ctx, u.ID, u.Email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	return token, model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (token string, user model.UserPublic, err error) {
	u, err := a.users.GetByEmail(ctx, email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	if u == nil || !auth.CheckPassword(u.PasswordHash, password) {
		return "", model.UserPublic{}, errs.ErrInvalidCredentials
	}

	_ = a.sessions.DeleteExpired(ctx, u.ID)

	token, err = a.mintToken(ctx, u.ID, u.Email)
	if err != nil {
		return "", model.UserPublic{}, err
	}
	return token, model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (a *Auth) Me(ctx context.Context, uid uuid.UUID) (model.UserPublic, error) {
	u, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return model.UserPublic{}, err
	}
	if u == nil {
		return model.UserPublic{}, errs.ErrNotFound
	}
	return model.UserPublic{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (a *Auth) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return a.sessions.Delete(ctx, sessionID)
}

func (a *Auth) ListUsers(ctx context.Context) ([]model.UserPublic, error) {
	return a.users.ListAllPublic(ctx, 500)
}

func (a *Auth) mintToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	expiresAt := time.Now().Add(a.tokenTTL)
	sessionID, err := a.sessions.Create(ctx, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return auth.SignToken(a.secret, userID, email, sessionID, expiresAt)
}
