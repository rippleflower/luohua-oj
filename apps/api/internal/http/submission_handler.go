package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type createSubmissionRequest struct {
	UserID    string `json:"userId"`
	ProblemID string `json:"problemId"`
	Language  string `json:"language"`
	Source    string `json:"source"`
}

type createSubmissionResponse struct {
	ID              string                            `json:"id"`
	UserID          string                            `json:"userId"`
	ProblemID       string                            `json:"problemId"`
	Problem         *submissionProblemSummaryResponse `json:"problem,omitempty"`
	Language        string                            `json:"language"`
	SourceObjectKey string                            `json:"sourceObjectKey"`
	Status          string                            `json:"status"`
	CreatedAt       string                            `json:"createdAt"`
}

type submissionDetailResponse struct {
	createSubmissionResponse
	CompileSummary       compileSummaryResponse       `json:"compileSummary"`
	Results              []submissionResultResponse   `json:"results"`
	ArtifactAvailability artifactAvailabilityResponse `json:"artifactAvailability"`
}

type submissionProblemSummaryResponse struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type listSubmissionsResponse struct {
	Items    []createSubmissionResponse `json:"items"`
	Total    int                        `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

type compileSummaryResponse struct {
	CompileOutput string `json:"compileOutput"`
	MaxTimeMs     *int32 `json:"maxTimeMs,omitempty"`
	MaxMemoryKb   *int32 `json:"maxMemoryKb,omitempty"`
	JudgedAt      string `json:"judgedAt,omitempty"`
}

type submissionResultResponse struct {
	TestCaseID    string `json:"testCaseId"`
	Status        string `json:"status"`
	TimeMs        *int32 `json:"timeMs,omitempty"`
	MemoryKb      *int32 `json:"memoryKb,omitempty"`
	OutputSnippet string `json:"outputSnippet,omitempty"`
	ErrorSnippet  string `json:"errorSnippet,omitempty"`
}

type artifactSummaryResponse struct {
	ArtifactType    string `json:"artifactType"`
	ObjectKey       string `json:"objectKey"`
	ContentType     string `json:"contentType,omitempty"`
	ContentEncoding string `json:"contentEncoding,omitempty"`
	ExpiresAt       string `json:"expiresAt,omitempty"`
}

type artifactAvailabilityResponse struct {
	SourceObjectKey string                    `json:"sourceObjectKey"`
	Artifacts       []artifactSummaryResponse `json:"artifacts"`
}

func createSubmissionHandler(creator submission.Creator, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.submissions")

	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		requestID := middleware.GetReqID(r.Context())

		handlerLogger.InfoContext(r.Context(),
			"submission request received",
			"event", "submission.request.received",
			"requestId", requestID,
		)

		var request createSubmissionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			handlerLogger.ErrorContext(r.Context(),
				"invalid json body",
				"event", "submission.request.invalid_json",
				"requestId", requestID,
				"error", err,
				"durationMs", time.Since(startedAt).Milliseconds(),
			)
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		userID, err := uuid.Parse(request.UserID)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(),
				"invalid userId",
				"event", "submission.request.invalid_user_id",
				"requestId", requestID,
				"error", err,
				"durationMs", time.Since(startedAt).Milliseconds(),
			)
			writeJSONError(w, http.StatusBadRequest, "invalid userId")
			return
		}
		problemID, err := uuid.Parse(request.ProblemID)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(),
				"invalid problemId",
				"event", "submission.request.invalid_problem_id",
				"requestId", requestID,
				"error", err,
				"durationMs", time.Since(startedAt).Milliseconds(),
			)
			writeJSONError(w, http.StatusBadRequest, "invalid problemId")
			return
		}

		created, err := creator.Create(r.Context(), submission.CreateInput{
			UserID:    userID,
			ProblemID: problemID,
			Language:  request.Language,
			Source:    request.Source,
		})
		if err != nil {
			handlerLogger.ErrorContext(r.Context(),
				"submission creation failed",
				"event", "submission.request.failed",
				"requestId", requestID,
				"error", err,
				"durationMs", time.Since(startedAt).Milliseconds(),
			)
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		handlerLogger.InfoContext(r.Context(),
			"submission request succeeded",
			"event", "submission.request.succeeded",
			"requestId", requestID,
			"submissionId", created.ID.String(),
			"durationMs", time.Since(startedAt).Milliseconds(),
		)

		writeJSON(w, http.StatusCreated, createSubmissionResponse{
			ID:              created.ID.String(),
			UserID:          created.UserID.String(),
			ProblemID:       created.ProblemID.String(),
			Problem:         problemSummaryResponseFromDomain(created.Problem),
			Language:        created.Language,
			SourceObjectKey: created.SourceObjectKey,
			Status:          string(created.Status),
			CreatedAt:       created.CreatedAt.Format(time.RFC3339),
		})
	}
}

func getSubmissionHandler(reader submission.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.submissions")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		submissionID, err := uuid.Parse(chi.URLParam(r, "submissionID"))
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "invalid submission id", "event", "submission.read.invalid_submission_id", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid submissionId")
			return
		}

		found, err := reader.Get(r.Context(), submissionID)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "submission read failed", "event", "submission.read.failed", "requestId", requestID, "submissionId", submissionID.String(), "error", err)
			writeJSONError(w, http.StatusNotFound, "submission not found")
			return
		}

		handlerLogger.InfoContext(r.Context(), "submission read succeeded", "event", "submission.read.succeeded", "requestId", requestID, "submissionId", found.ID.String())
		writeJSON(w, http.StatusOK, submissionDetailResponse{
			createSubmissionResponse: createSubmissionResponse{
				ID:              found.ID.String(),
				UserID:          found.UserID.String(),
				ProblemID:       found.ProblemID.String(),
				Problem:         problemSummaryResponseFromDomain(found.Problem),
				Language:        found.Language,
				SourceObjectKey: found.SourceObjectKey,
				Status:          string(found.Status),
				CreatedAt:       found.CreatedAt.Format(time.RFC3339),
			},
			CompileSummary:       compileSummaryResponseFromDomain(found.CompileSummary),
			Results:              submissionResultsResponseFromDomain(found.Results),
			ArtifactAvailability: artifactAvailabilityResponseFromDomain(found.ArtifactAvailability),
		})
	}
}

func listUserSubmissionsHandler(reader submission.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.submissions")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		username := chi.URLParam(r, "username")
		page, err := parsePositiveIntQuery(r, "page")
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "invalid page", "event", "submission.list.invalid_page", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid page")
			return
		}
		pageSize, err := parsePositiveIntQuery(r, "pageSize")
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "invalid page size", "event", "submission.list.invalid_page_size", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid pageSize")
			return
		}

		submissions, err := reader.ListByUsername(r.Context(), submission.ListByUsernameParams{
			Username: username,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "user submissions list failed", "event", "submission.list.failed", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		response := make([]createSubmissionResponse, 0, len(submissions.Items))
		for _, item := range submissions.Items {
			response = append(response, createSubmissionResponse{
				ID:              item.ID.String(),
				UserID:          item.UserID.String(),
				ProblemID:       item.ProblemID.String(),
				Problem:         problemSummaryResponseFromDomain(item.Problem),
				Language:        item.Language,
				SourceObjectKey: item.SourceObjectKey,
				Status:          string(item.Status),
				CreatedAt:       item.CreatedAt.Format(time.RFC3339),
			})
		}

		handlerLogger.InfoContext(r.Context(), "user submissions listed", "event", "submission.list.succeeded", "requestId", requestID, "username", username, "count", len(response), "page", submissions.Page, "pageSize", submissions.PageSize, "total", submissions.Total)
		writeJSON(w, http.StatusOK, listSubmissionsResponse{
			Items:    response,
			Total:    submissions.Total,
			Page:     submissions.Page,
			PageSize: submissions.PageSize,
		})
	}
}

func parsePositiveIntQuery(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, strconv.ErrSyntax
	}

	return value, nil
}

func problemSummaryResponseFromDomain(problem *submission.ProblemSummary) *submissionProblemSummaryResponse {
	if problem == nil {
		return nil
	}

	return &submissionProblemSummaryResponse{
		ID:    problem.ID.String(),
		Slug:  problem.Slug,
		Title: problem.Title,
	}
}

func compileSummaryResponseFromDomain(summary submission.CompileSummary) compileSummaryResponse {
	response := compileSummaryResponse{
		CompileOutput: summary.CompileOutput,
		MaxTimeMs:     summary.MaxTimeMs,
		MaxMemoryKb:   summary.MaxMemoryKb,
	}
	if summary.JudgedAt != nil {
		response.JudgedAt = summary.JudgedAt.UTC().Format(time.RFC3339)
	}
	return response
}

func submissionResultsResponseFromDomain(results []submission.ResultSummary) []submissionResultResponse {
	response := make([]submissionResultResponse, 0, len(results))
	for _, item := range results {
		response = append(response, submissionResultResponse{
			TestCaseID:    item.TestCaseID.String(),
			Status:        string(item.Status),
			TimeMs:        item.TimeMs,
			MemoryKb:      item.MemoryKb,
			OutputSnippet: item.OutputSnippet,
			ErrorSnippet:  item.ErrorSnippet,
		})
	}
	return response
}

func artifactAvailabilityResponseFromDomain(availability submission.ArtifactAvailability) artifactAvailabilityResponse {
	response := artifactAvailabilityResponse{
		SourceObjectKey: availability.SourceObjectKey,
		Artifacts:       make([]artifactSummaryResponse, 0, len(availability.Artifacts)),
	}
	for _, item := range availability.Artifacts {
		artifact := artifactSummaryResponse{
			ArtifactType:    item.ArtifactType,
			ObjectKey:       item.ObjectKey,
			ContentType:     item.ContentType,
			ContentEncoding: item.ContentEncoding,
		}
		if item.ExpiresAt != nil {
			artifact.ExpiresAt = item.ExpiresAt.UTC().Format(time.RFC3339)
		}
		response.Artifacts = append(response.Artifacts, artifact)
	}
	return response
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
