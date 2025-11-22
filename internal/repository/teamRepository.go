package repository

import (
	"context"
	"errors"

	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamRepository struct {
	db DBExecutor
}

func NewTeamRepository(db DBExecutor) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateTeam создает команду. Если команда существует - возвращаем ошибку
func (r *TeamRepository) CreateTeam(ctx context.Context, name string) error {
	query := `INSERT INTO teams (team_name) VALUES ($1)`

	if _, err := r.db.Exec(ctx, query, name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ErrTeamExists
		}
		return err
	}
	return nil
}

// CheckTeamExists проверяет, есть ли команда
func (r *TeamRepository) CheckTeamExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	return exists, err
}
