package auth

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidCredential = errors.New("invalid username or password")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, dto RegisterDTO) (*AuthResponse, error) {
	_, err := s.repo.FindByUsername(ctx, dto.Username)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Insert(ctx, &Auth{
		Username:     dto.Username,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

func (s *Service) Login(ctx context.Context, dto LoginDTO) (*AuthResponse, error) {
	user, err := s.repo.FindByUsername(ctx, dto.Username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		return nil, ErrInvalidCredential
	}

	return toResponse(user), nil
}

func toResponse(user *Auth) *AuthResponse {
	return &AuthResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}
}
