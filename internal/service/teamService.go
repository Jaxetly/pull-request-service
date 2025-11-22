package service

import (
	"context"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/Jaxetly/pull-request-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamService struct {
	pool *pgxpool.Pool
}

func NewTeamService(pool *pgxpool.Pool) *TeamService {
	return &TeamService{pool: pool}
}

// CreateTeam создает команду и сразу добавляет в неё участников
func (s *TeamService) CreateTeam(ctx context.Context, team api.Team) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	teamRep := repository.NewTeamRepository(tx)

	if err := teamRep.CreateTeam(ctx, team.TeamName); err != nil {
		return err
	}

	userRep := repository.NewUserRepository(tx)

	for _, teamUser := range team.Members {
		user := api.User{
			UserId:   teamUser.UserId,
			Username: teamUser.Username,
			TeamName: team.TeamName,
			IsActive: teamUser.IsActive,
		}

		if err := userRep.UpsertUser(ctx, user); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetTeam возвращает полную информацию о команде, включая список участников
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (api.Team, error) {
	returnError := func(err error) (api.Team, error) {
		return api.Team{}, err
	}

	teamRep := repository.NewTeamRepository(s.pool)

	if exists, err := teamRep.CheckTeamExists(ctx, teamName); err != nil {
		return returnError(err)
	} else if !exists {
		return returnError(errs.ErrTeamNotFound)
	}

	team := api.Team{TeamName: teamName}

	userRep := repository.NewUserRepository(s.pool)

	teamMembers, err := userRep.GetUsersFromTeam(ctx, teamName)
	if err != nil {
		return returnError(err)
	}

	team.Members = teamMembers

	return team, nil
}
