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
	returnError := func(err error) (api.User, error) {
		return api.User{}, err
	}

	userRep := repository.NewUserRepository(s.pool)

	user, err := userRep.GetUser(ctx, userID)
	if err != nil {
		return api.User{}, err
	}

	if user.IsActive == isActive {
		// ничего не делаем
		return user, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return returnError(err)
	}
	defer tx.Rollback(ctx)

	userRep = repository.NewUserRepository(tx)

	if err := userRep.SetActive(ctx, userID, isActive); err != nil {
		return api.User{}, err
	}

	user, err = userRep.GetUser(ctx, userID)
	if err != nil {
		return api.User{}, err
	}

	prRep := repository.NewPRRepository(tx)

	if !isActive {
		prs, err := prRep.GetUserReviews(ctx, user.UserId)
		if err != nil {
			return returnError(err)
		}

		for _, pr := range prs {
			// работаем только с открытыми PR
			if pr.Status != api.PullRequestShortStatusOPEN {
				continue
			}

			candidates, err := userRep.GetRandomActiveUsersFromTeam(ctx, user.TeamName, 1, []string{user.UserId, pr.AuthorId})
			if err != nil {
				return returnError(err)
			}

			if err := prRep.RemoveReviewer(ctx, pr.PullRequestId, user.UserId); err != nil {
				return returnError(err)
			}

			if len(candidates) > 0 {
				// Нашли замену — подменяем ревьювера.
				if err := prRep.AddReviewers(ctx, pr.PullRequestId, candidates); err != nil {
					return returnError(err)
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return returnError(err)
	}

	return user, nil
}

// DeactivateUsersByTeam деактивирует всех пользователей команды
func (s *UserService) DeactivateUsersByTeam(ctx context.Context, teamName string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	teamRep := repository.NewTeamRepository(tx)

	exists, err := teamRep.CheckTeamExists(ctx, teamName)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errs.ErrTeamNotFound
	}

	userRep := repository.NewUserRepository(tx)

	deactivatedUsers, err := userRep.DeactivateUsersByTeam(ctx, teamName)
	if err != nil {
		return 0, err
	}

	prRep := repository.NewPRRepository(tx)

	if err := prRep.RemoveReviewersByTeam(ctx, teamName); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return deactivatedUsers, nil
}
