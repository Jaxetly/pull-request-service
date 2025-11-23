package service

import (
	"context"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/Jaxetly/pull-request-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool *pgxpool.Pool
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{pool: pool}
}

// GetUserReviews возвращает список PR, где пользователь является ревьювером
func (s *UserService) GetUserReviews(ctx context.Context, userID string) ([]api.PullRequestShort, error) {
	userRep := repository.NewUserRepository(s.pool)

	exists, err := userRep.CheckUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotFound
	}

	prRep := repository.NewPRRepository(s.pool)

	result, err := prRep.GetUserReviews(ctx, userID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SetActive меняет статус активности пользователя (is_active)
func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (api.User, error) {
	userRep := repository.NewUserRepository(s.pool)

	if err := userRep.SetActive(ctx, userID, isActive); err != nil {
		return api.User{}, err
	}

	user, err := userRep.GetUser(ctx, userID)
	if err != nil {
		return api.User{}, err
	}

	return user, nil
}
