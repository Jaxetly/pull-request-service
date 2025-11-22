package errs

import "errors"

var (
	ErrTeamNotFound        = errors.New("team not found")
	ErrTeamExists          = errors.New("team already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRMerged            = errors.New("pull request is already merged")
	ErrNoCandidate         = errors.New("there are no available candidates for this pull request")
	ErrReviewerNotAssigned = errors.New("user is not assigned as reviewer for this pull request")
)
