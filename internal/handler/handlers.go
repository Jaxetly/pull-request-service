package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/customtypes"
	"github.com/Jaxetly/pull-request-service/internal/errs"
	"github.com/Jaxetly/pull-request-service/internal/service"
)

// Server реализует api.ServerInterface
type Server struct {
	teamService *service.TeamService
	userService *service.UserService
	prService   *service.PRService
}

func NewServer(team *service.TeamService, user *service.UserService, pr *service.PRService) *Server {
	return &Server{
		teamService: team,
		userService: user,
		prService:   pr,
	}
}

func (s *Server) mapError(err error) (int, api.ErrorResponse) {
	var resp api.ErrorResponse

	setErrorResponse := func(code api.ErrorResponseErrorCode, msg string) api.ErrorResponse {
		resp.Error.Code = code
		resp.Error.Message = msg
		return resp
	}

	switch {
	case errors.Is(err, errs.ErrTeamExists):
		return http.StatusBadRequest, setErrorResponse(api.TEAMEXISTS, err.Error())

	case errors.Is(err, errs.ErrPRExists):
		return http.StatusConflict, setErrorResponse(api.PREXISTS, err.Error())

	case errors.Is(err, errs.ErrPRMerged):
		return http.StatusConflict, setErrorResponse(api.PRMERGED, err.Error())

	case errors.Is(err, errs.ErrNoCandidate):
		return http.StatusConflict, setErrorResponse(api.NOCANDIDATE, err.Error())

	case errors.Is(err, errs.ErrReviewerNotAssigned):
		return http.StatusConflict, setErrorResponse(api.NOTASSIGNED, err.Error())

	case errors.Is(err, errs.ErrTeamNotFound),
		errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrPRNotFound):
		return http.StatusNotFound, setErrorResponse(api.NOTFOUND, err.Error())

	default:
		return http.StatusInternalServerError, setErrorResponse(api.INTERNALERROR, err.Error())
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if v == nil {
		return
	}

	json.NewEncoder(w).Encode(v)
}

func (s *Server) writeErrorResponse(w http.ResponseWriter, status int, resp api.ErrorResponse) {
	s.writeJSON(w, status, resp)
}

func (s *Server) writeErrorResponseBadRequest(w http.ResponseWriter, err error) {
	var resp api.ErrorResponse
	resp.Error.Code = api.BADREQUEST
	resp.Error.Message = err.Error()

	s.writeJSON(w, http.StatusBadRequest, resp)
}

// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// (POST /pullRequest/create)
func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeErrorResponseBadRequest(w, err)
		return
	}

	pr, err := s.prService.CreatePR(ctx, api.PostPullRequestCreateJSONBody(body))
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	response := customtypes.PostPullRequestResponse{
		PR: pr,
	}
	s.writeJSON(w, http.StatusCreated, response)
}

// Пометить PR как MERGED (идемпотентная операция)
// (POST /pullRequest/merge)
func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeErrorResponseBadRequest(w, err)
		return
	}

	pr, err := s.prService.MergePR(ctx, body.PullRequestId)
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	response := customtypes.PostPullRequestResponse{
		PR: pr,
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Переназначить конкретного ревьювера на другого из его команды
// (POST /pullRequest/reassign)
func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeErrorResponseBadRequest(w, err)
		return
	}

	pr, newUserID, err := s.prService.ReassignReviewer(ctx, body.PullRequestId, body.OldUserId)
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	response := customtypes.PostPullRequestReassignResponse{
		PR:         pr,
		ReplacedBy: newUserID,
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Статистика по всем командам
// (GET /stats/teams)
func (s *Server) GetStatsTeams(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Статистика по пользователям (ревью и открытые PR)
// (GET /stats/users)
func (s *Server) GetStatsUsers(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Создать команду с участниками (создаёт/обновляет пользователей)
// (POST /team/add)
func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeErrorResponseBadRequest(w, err)
		return
	}

	if err := s.teamService.CreateTeam(ctx, body); err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	response := customtypes.PostTeamResponse{
		Team: body,
	}
	s.writeJSON(w, http.StatusCreated, response)
}

// Массовая деактивация пользователей команды
// (POST /team/deactivateUsers)
func (s *Server) PostTeamDeactivateUsers(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Получить команду с участниками
// (GET /team/get)
func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	ctx := r.Context()

	team, err := s.teamService.GetTeam(ctx, params.TeamName)
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	s.writeJSON(w, http.StatusOK, team)
}

// Получить PR'ы, где пользователь назначен ревьювером
// (GET /users/getReview)
func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	ctx := r.Context()

	reviews, err := s.userService.GetUserReviews(ctx, params.UserId)
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	result := customtypes.GetUsersGetReviewResponse{
		UserID:       params.UserId,
		PullRequests: reviews,
	}
	s.writeJSON(w, http.StatusOK, result)
}

// Установить флаг активности пользователя
// (POST /users/setIsActive)
func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeErrorResponseBadRequest(w, err)
		return
	}

	user, err := s.userService.SetActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		status, resp := s.mapError(err)
		s.writeErrorResponse(w, status, resp)
		return
	}

	response := customtypes.PostUsersResponse{
		User: user,
	}
	s.writeJSON(w, http.StatusOK, response)
}
