package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/google/uuid"
)

var (
	errInvalidStartsAt   = errors.New("invalid startsAt")
	errInvalidEndsAt     = errors.New("invalid endsAt")
	errInvalidProblemID  = errors.New("invalid problem id")
	errInvalidContestID  = errors.New("invalid contest id")
	errInvalidUserID     = errors.New("invalid user id")
	errInvalidSubmission = errors.New("invalid submission id")
	errInvalidAnnID      = errors.New("invalid announcement id")
)

type adminUserUpdateRequest struct {
	Status      string `json:"status"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatarUrl"`
	Reason      string `json:"reason"`
}

type adminUserRoleRequest struct {
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

type adminUserPermissionsUpdateRequest struct {
	Permissions []string `json:"permissions"`
	Reason      string   `json:"reason"`
}

type adminProblemUpsertRequest struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Difficulty    string `json:"difficulty"`
	TimeLimitMs   int    `json:"timeLimitMs"`
	MemoryLimitKb int    `json:"memoryLimitKb"`
	Reason        string `json:"reason"`
}

type adminReasonOnlyRequest struct {
	Reason string `json:"reason"`
}

type adminContestUpsertRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	StartsAt    string `json:"startsAt"`
	EndsAt      string `json:"endsAt"`
	Reason      string `json:"reason"`
}

type adminContestProblemBindingRequest struct {
	ProblemID string `json:"problemId"`
	Code      string `json:"code"`
	Position  int    `json:"position"`
}

type adminContestProblemsReplaceRequest struct {
	Problems []adminContestProblemBindingRequest `json:"problems"`
	Reason   string                              `json:"reason"`
}

type adminSystemSettingsUpdateRequest struct {
	RegistrationEnabled bool   `json:"registrationEnabled"`
	JudgeQueuePaused    bool   `json:"judgeQueuePaused"`
	StorageMode         string `json:"storageMode"`
}

type adminAnnouncementUpsertRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Status   string `json:"status"`
	Audience string `json:"audience"`
}

func decodeAdminUserUpdateRequest(r *http.Request) (adminUserUpdateRequest, error) {
	var request adminUserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminUserUpdateRequest{}, err
	}
	return request, nil
}

func decodeAdminUserRoleRequest(r *http.Request) (adminUserRoleRequest, error) {
	var request adminUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminUserRoleRequest{}, err
	}
	return request, nil
}

func decodeAdminUserPermissionsUpdateRequest(r *http.Request) (adminUserPermissionsUpdateRequest, error) {
	var request adminUserPermissionsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminUserPermissionsUpdateRequest{}, err
	}
	return request, nil
}

func decodeAdminProblemUpsertRequest(r *http.Request) (adminProblemUpsertRequest, error) {
	var request adminProblemUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminProblemUpsertRequest{}, err
	}
	return request, nil
}

func decodeAdminReasonOnlyRequest(r *http.Request) (adminReasonOnlyRequest, error) {
	var request adminReasonOnlyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminReasonOnlyRequest{}, err
	}
	return request, nil
}

func decodeAdminContestUpsertRequest(r *http.Request) (adminContestUpsertRequest, error) {
	var request adminContestUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminContestUpsertRequest{}, err
	}
	return request, nil
}

func decodeAdminContestProblemsReplaceRequest(r *http.Request) (adminContestProblemsReplaceRequest, error) {
	var request adminContestProblemsReplaceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminContestProblemsReplaceRequest{}, err
	}
	return request, nil
}

func decodeAdminSystemSettingsUpdateRequest(r *http.Request) (adminSystemSettingsUpdateRequest, error) {
	var request adminSystemSettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminSystemSettingsUpdateRequest{}, err
	}
	return request, nil
}

func decodeAdminAnnouncementUpsertRequest(r *http.Request) (adminAnnouncementUpsertRequest, error) {
	var request adminAnnouncementUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return adminAnnouncementUpsertRequest{}, err
	}
	return request, nil
}

func parseAdminUserID(r *http.Request) (uuid.UUID, error) {
	userID, err := parseUUIDURLParam(r, "userID")
	if err != nil {
		return uuid.Nil, errInvalidUserID
	}
	return userID, nil
}

func parseAdminProblemID(r *http.Request) (uuid.UUID, error) {
	problemID, err := parseUUIDURLParam(r, "problemID")
	if err != nil {
		return uuid.Nil, errInvalidProblemID
	}
	return problemID, nil
}

func parseAdminContestID(r *http.Request) (uuid.UUID, error) {
	contestID, err := parseUUIDURLParam(r, "contestID")
	if err != nil {
		return uuid.Nil, errInvalidContestID
	}
	return contestID, nil
}

func parseAdminSubmissionID(r *http.Request) (uuid.UUID, error) {
	submissionID, err := parseUUIDURLParam(r, "submissionID")
	if err != nil {
		return uuid.Nil, errInvalidSubmission
	}
	return submissionID, nil
}

func parseAdminAnnouncementID(r *http.Request) (uuid.UUID, error) {
	announcementID, err := parseUUIDURLParam(r, "announcementID")
	if err != nil {
		return uuid.Nil, errInvalidAnnID
	}
	return announcementID, nil
}

func parseAdminContestSchedule(request adminContestUpsertRequest) (time.Time, time.Time, error) {
	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, errInvalidStartsAt
	}
	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		return time.Time{}, time.Time{}, errInvalidEndsAt
	}
	return startsAt, endsAt, nil
}

func parseAdminContestProblemBindings(request adminContestProblemsReplaceRequest) ([]contest.AdminProblemBindingInput, error) {
	problems := make([]contest.AdminProblemBindingInput, 0, len(request.Problems))
	for _, item := range request.Problems {
		problemID, err := uuid.Parse(item.ProblemID)
		if err != nil {
			return nil, errInvalidProblemID
		}
		problems = append(problems, contest.AdminProblemBindingInput{
			ProblemID: problemID,
			Code:      item.Code,
			Position:  item.Position,
		})
	}
	return problems, nil
}
