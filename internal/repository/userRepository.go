package repository

import (
	"context"
	"errors"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db DBExecutor
}

func NewUserRepository(db DBExecutor) *UserRepository {
	return &UserRepository{db: db}
}

// GetUser получает пользователя
func (r *UserRepository) GetUser(ctx context.Context, userID string) (api.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	var user api.User
	if err := r.db.QueryRow(ctx, query, userID).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive); err != nil {
		return api.User{}, err
	}
	return user, nil
}

// UpsertUser создает или обновляет пользователя
func (r *UserRepository) UpsertUser(ctx context.Context, user api.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active
	`
	_, err := r.db.Exec(ctx, query, user.UserId, user.Username, user.TeamName, user.IsActive)
	return err
}

// GetUsersFromTeam получает всех пользователей состоящих в команде
func (r *UserRepository) GetUsersFromTeam(ctx context.Context, teamName string) ([]api.TeamMember, error) {
	query := `SELECT user_id, username, is_active FROM users WHERE team_name = $1`
	rows, err := r.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []api.TeamMember{}
	for rows.Next() {
		var m api.TeamMember
		if err := rows.Scan(&m.UserId, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

// SetActive устанавливает активность пользователя
func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE user_id = $2`
	tag, err := r.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrUserNotFound
	}
	return nil
}

// GetRandomActiveUsersFromTeam получает <= userCount случайных кандидатов из команды, исключая список (например, автора или уже назначенного)
func (r *UserRepository) GetRandomActiveUsersFromTeam(ctx context.Context, teamName string, userCount uint, excludeUserIDs []string) ([]string, error) {
	query := `
		SELECT user_id FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != ALL($2)
		ORDER BY RANDOM() 
		LIMIT $3
	`
	var usersIDs []string
	rows, err := r.db.Query(ctx, query, teamName, excludeUserIDs, userCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		usersIDs = append(usersIDs, userID)
	}

	return usersIDs, nil
}

// GetUserTeam получает имя команды пользователя
func (r *UserRepository) GetUserTeam(ctx context.Context, userID string) (string, error) {
	var teamName string
	if err := r.db.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id = $1`, userID).Scan(&teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrUserNotFound
		}
		return "", err
	}
	return teamName, nil
}

// DeactivateUsersByTeam деактивирует всех пользователей команды
func (r *UserRepository) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	query := `UPDATE users SET is_active = false WHERE team_name = $1`
	_, err := r.db.Exec(ctx, query, teamName)
	return err
}

// CheckUserExists проверяет, есть ли пользователь
func (r *UserRepository) CheckUserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	return exists, err
}
