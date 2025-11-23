package customtypes

import "github.com/Jaxetly/pull-request-service/internal/api"

type PostTeamResponse struct {
	Team api.Team `json:"team"`
}

type PostUsersResponse struct {
	User api.User `json:"user"`
}

type GetUsersGetReviewResponse struct {
	UserID       string                 `json:"user_id"`
	PullRequests []api.PullRequestShort `json:"pull_requests"`
}

type PostPullRequestResponse struct {
	PR api.PullRequest `json:"pr"`
}

type PostPullRequestReassignResponse struct {
	PR         api.PullRequest `json:"pr"`
	ReplacedBy string          `json:"replaced_by"`
}
