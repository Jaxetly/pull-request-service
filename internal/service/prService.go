package service

import (
	"context"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/Jaxetly/pull-request-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRService struct {
	pool *pgxpool.Pool
}

func NewPRService(pool *pgxpool.Pool) *PRService {
	return &PRService{pool: pool}
}

// CreatePR создает новый PR и назначает ревьюверов
func (s *PRService) CreatePR(ctx context.Context, req api.PostPullRequestCreateJSONBody) (api.PullRequest, error) {
	returnError := func(err error) (api.PullRequest, error) {
		return api.PullRequest{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return returnError(err)
	}
	defer tx.Rollback(ctx)

	prRep := repository.NewPRRepository(tx)
	userRep := repository.NewUserRepository(tx)

	pr := api.PullRequest{
		PullRequestId:   req.PullRequestId,
		PullRequestName: req.PullRequestName,
		AuthorId:        req.AuthorId,
	}

	if exists, err := userRep.CheckUserExists(ctx, req.AuthorId); err != nil {
		return returnError(err)
	} else if !exists {
		return returnError(errs.ErrUserNotFound)
	}

	if err := prRep.CreatePR(ctx, pr); err != nil {
		return returnError(err)
	}

	teamName, err := userRep.GetUserTeam(ctx, pr.AuthorId)
	if err != nil {
		return returnError(err)
	}

	reviewersIDs, err := userRep.GetRandomActiveUsersFromTeam(ctx, teamName, 2, []string{pr.AuthorId})
	if err != nil {
		return returnError(err)
	}

	if err := prRep.AddReviewers(ctx, pr.PullRequestId, reviewersIDs); err != nil {
		return returnError(err)
	}

	newPr, err := prRep.GetPR(ctx, pr.PullRequestId)
	if err != nil {
		return returnError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return returnError(err)
	}

	return newPr, nil
}

// MergePR переводит PR в статус MERGED
func (s *PRService) MergePR(ctx context.Context, prID string) (api.PullRequest, error) {
	returnError := func(err error) (api.PullRequest, error) {
		return api.PullRequest{}, err
	}

	prRep := repository.NewPRRepository(s.pool)

	if exists, err := prRep.CheckPRExists(ctx, prID); err != nil {
		return returnError(err)
	} else if !exists {
		return returnError(errs.ErrPRNotFound)
	}

	if err := prRep.MergePR(ctx, prID); err != nil {
		return returnError(err)
	}

	pr, err := prRep.GetPR(ctx, prID)
	if err != nil {
		return returnError(err)
	}

	return pr, nil
}

// ReassignReviewer меняет одного ревьювера на другого
func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (newPr api.PullRequest, newUserID string, err error) {
	returnError := func(err error) (api.PullRequest, string, error) {
		return api.PullRequest{}, "", err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return returnError(err)
	}
	defer tx.Rollback(ctx)

	prRep := repository.NewPRRepository(tx)

	pr, err := prRep.GetPR(ctx, prID)
	if err != nil {
		return returnError(err)
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return returnError(errs.ErrPRMerged)
	}

	if !contains(pr.AssignedReviewers, oldUserID) {
		return returnError(errs.ErrReviewerNotAssigned)
	}

	if err := prRep.RemoveReviewer(ctx, prID, oldUserID); err != nil {
		return returnError(err)
	}

	userRep := repository.NewUserRepository(tx)

	teamName, err := userRep.GetUserTeam(ctx, oldUserID)
	if err != nil {
		return returnError(err)
	}

	reviewersIDs, err := userRep.GetRandomActiveUsersFromTeam(ctx, teamName, 1, []string{pr.AuthorId, oldUserID})
	if err != nil {
		return returnError(err)
	}

	if len(reviewersIDs) == 0 {
		return returnError(errs.ErrNoCandidate)
	}

	if err := prRep.AddReviewers(ctx, pr.PullRequestId, reviewersIDs); err != nil {
		return returnError(err)
	}

	newPr, err = prRep.GetPR(ctx, prID)
	if err != nil {
		return returnError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return returnError(err)
	}

	return newPr, reviewersIDs[0], nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
