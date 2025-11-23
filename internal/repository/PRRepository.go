package repository

import (
	"context"
	"errors"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PRRepository struct {
	db DBExecutor
}

func NewPRRepository(db DBExecutor) *PRRepository {
	return &PRRepository{db: db}
}

// CreatePR создает запись о PR.
// Важно: ревьюверов мы добавляем отдельным запросом в pr_reviewers
func (r *PRRepository) CreatePR(ctx context.Context, pr api.PullRequest) error {
	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.db.Exec(ctx, query, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, api.PullRequestStatusOPEN)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ErrPRExists
		}
		return err
	}

	return nil
}

// AddReviewers массово добавляет ревьюверов
func (r *PRRepository) AddReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}

	query := `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
	for _, userID := range reviewerIDs {
		if _, err := r.db.Exec(ctx, query, prID, userID); err != nil {
			return err
		}
	}
	return nil
}

// GetPR получает PR по ID
func (r *PRRepository) GetPR(ctx context.Context, prID string) (api.PullRequest, error) {
	returnError := func(err error) (api.PullRequest, error) {
		return api.PullRequest{}, err
	}

	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1
	`
	var pr api.PullRequest
	var statusStr string

	err := r.db.QueryRow(ctx, query, prID).Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &statusStr, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return returnError(errs.ErrPRNotFound)
		}
		return returnError(err)
	}

	pr.Status = api.PullRequestStatus(statusStr)

	reviewersQuery := `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1`

	rows, err := r.db.Query(ctx, reviewersQuery, prID)
	if err != nil {
		return returnError(err)
	}
	defer rows.Close()

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return returnError(err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return returnError(err)
	}

	return pr, nil
}

// MergePR обновляет статус. Идемпотентность обеспечивается условием WHERE.
func (r *PRRepository) MergePR(ctx context.Context, prID string) error {
	query := `UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE pull_request_id = $2 AND status != $1`

	_, err := r.db.Exec(ctx, query, api.PullRequestStatusMERGED, prID)
	return err
}

// RemoveReviewer удаляет конкретного ревьювера
func (r *PRRepository) RemoveReviewer(ctx context.Context, prID, userID string) error {
	query := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`

	_, err := r.db.Exec(ctx, query, prID, userID)
	return err
}

// RemoveReviewersByTeam удаляет всех ревьюверов из указанной команды из любых PR
func (r *PRRepository) RemoveReviewersByTeam(ctx context.Context, teamName string) error {
	query := `
		DELETE FROM pr_reviewers
		WHERE user_id IN (
			SELECT u.user_id 
			FROM users u 
			WHERE u.team_name = $1
		)
		AND pull_request_id IN (
			SELECT pull_request_id 
			FROM pull_requests 
			WHERE status = $2
		)
	`
	_, err := r.db.Exec(ctx, query, teamName, api.PullRequestStatusOPEN)
	return err
}

// GetUserReviews возвращает список PR, где юзер является ревьювером
func (r *PRRepository) GetUserReviews(ctx context.Context, userID string) ([]api.PullRequestShort, error) {
	query := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pr_reviewers rv
		JOIN pull_requests pr USING(pull_request_id)
		WHERE rv.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []api.PullRequestShort{}
	for rows.Next() {
		var prShort api.PullRequestShort
		var statusStr string
		err := rows.Scan(&prShort.PullRequestId, &prShort.PullRequestName, &prShort.AuthorId, &statusStr)
		if err != nil {
			return nil, err
		}
		prShort.Status = api.PullRequestShortStatus(statusStr)

		result = append(result, prShort)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// CheckPRExists проверяет, есть ли pull request
func (r *PRRepository) CheckPRExists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	err := r.db.QueryRow(ctx, query, prID).Scan(&exists)
	return exists, err
}
