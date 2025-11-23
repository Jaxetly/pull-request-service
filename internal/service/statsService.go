package service

import (
	"context"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsService struct {
	pool *pgxpool.Pool
}

func NewStatsService(pool *pgxpool.Pool) *StatsService {
	return &StatsService{pool: pool}
}

func (s *StatsService) GetTeamStats(ctx context.Context) (api.TeamStatsResponse, error) {
	statsRep := repository.NewStatsRepository(s.pool)

	items, err := statsRep.GetTeamStats(ctx)
	if err != nil {
		return api.TeamStatsResponse{}, err
	}

	stats := api.TeamStatsResponse{
		Items: items,
	}

	return stats, nil
}

func (s *StatsService) GetUserStats(ctx context.Context) (api.UserStatsResponse, error) {
	statsRep := repository.NewStatsRepository(s.pool)

	items, err := statsRep.GetUserStats(ctx)
	if err != nil {
		return api.UserStatsResponse{}, err
	}

	stats := api.UserStatsResponse{
		Items: items,
	}

	return stats, nil
}
