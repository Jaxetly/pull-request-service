package errs

import "errors"

var (
	ErrTeamNotFound        = errors.New("team not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrPRNotFound          = errors.New("PR not found")
	ErrTeamExists          = errors.New("team_name already exists")
	ErrPRExists            = errors.New("PR id already exists")
	ErrPRMerged            = errors.New("cannot reassign on merged PR")
	ErrNoCandidate         = errors.New("no active replacement candidate in team")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
)
