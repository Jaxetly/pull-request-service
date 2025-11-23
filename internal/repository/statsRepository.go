package repository

import (
	"context"

	"github.com/Jaxetly/pull-request-service/internal/api"
)

type StatsRepository struct {
	db DBExecutor
}

func NewStatsRepository(db DBExecutor) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetTeamStats(ctx context.Context) ([]api.TeamStatsItem, error) {
	query := `
		SELECT 
			t.team_name,
			COUNT(u.user_id) as members_count,
			COUNT(u.user_id) FILTER (WHERE u.is_active = true) as active_members_count
		FROM teams t
		LEFT JOIN users u USING(team_name)
		GROUP BY t.team_name
		ORDER BY t.team_name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []api.TeamStatsItem
	for rows.Next() {
		var item api.TeamStatsItem
		if err := rows.Scan(&item.TeamName, &item.MembersCount, &item.ActiveMembersCount); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *StatsRepository) GetUserStats(ctx context.Context) ([]api.UserStatsItem, error) {
	query := `
		SELECT
		u.user_id,
		u.team_name,
		COALESCE(
			(
				SELECT COUNT(*)
				FROM pr_reviewers prr
				JOIN pull_requests pr ON prr.pull_request_id = pr.pull_request_id
				WHERE prr.user_id = u.user_id AND pr.status = $1
			), 0
		) as reviews_count,
		COALESCE(
			(
				SELECT COUNT(*)
				FROM pull_requests pr_authored
				WHERE pr_authored.author_id = u.user_id AND pr_authored.status = $1
			), 0
		) as open_authored_prs
		FROM users u
		ORDER BY u.user_id
	`

	rows, err := r.db.Query(ctx, query, api.PullRequestShortStatusOPEN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []api.UserStatsItem
	for rows.Next() {
		var item api.UserStatsItem
		if err := rows.Scan(&item.UserId, &item.TeamName, &item.ReviewsCount, &item.OpenAuthoredPrs); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
