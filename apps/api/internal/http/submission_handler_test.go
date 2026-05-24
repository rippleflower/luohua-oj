package http_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateSubmission(t *testing.T) {
	userID := uuid.New()
	problemID := uuid.New()
	creator := &fakeSubmissionCreator{
		submission: submission.Submission{
			ID:              uuid.New(),
			UserID:          userID,
			ProblemID:       problemID,
			Language:        "CPP17",
			SourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
			Status:          submission.StatusPending,
			CreatedAt:       time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	body := []byte(`{
		"userId":"` + userID.String() + `",
		"problemId":"` + problemID.String() + `",
		"language":"CPP17",
		"source":"int main(){return 0;}"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{
		SubmissionCreator: creator,
		Logger:            slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
	}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{
		"id":"`+creator.submission.ID.String()+`",
		"userId":"`+userID.String()+`",
		"problemId":"`+problemID.String()+`",
		"language":"CPP17",
		"sourceObjectKey":"submissions/2026/05/sub-1/source.cpp.zst",
		"status":"PENDING",
		"createdAt":"2026-05-15T00:00:00Z"
	}`, rec.Body.String())
}

func TestCreateSubmissionRejectsBadProblemID(t *testing.T) {
	body := []byte(`{
		"userId":"` + uuid.New().String() + `",
		"problemId":"bad-id",
		"language":"CPP17",
		"source":"int main(){return 0;}"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	var logBuffer bytes.Buffer

	apphttp.NewRouter(apphttp.RouterOptions{
		SubmissionCreator: &fakeSubmissionCreator{},
		Logger:            slog.New(slog.NewJSONHandler(&logBuffer, nil)),
	}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{"error":"invalid problemId"}`, rec.Body.String())
	require.True(t, strings.Contains(logBuffer.String(), `"event":"submission.request.invalid_problem_id"`))
}

type fakeSubmissionCreator struct {
	submission submission.Submission
}

func (f *fakeSubmissionCreator) Create(ctx context.Context, input submission.CreateInput) (submission.Submission, error) {
	return f.submission, nil
}

type fakeSubmissionReader struct {
	submissionDetail submission.SubmissionDetail
	page             submission.SubmissionPage
}

func (f *fakeSubmissionReader) Get(ctx context.Context, submissionID uuid.UUID) (submission.SubmissionDetail, error) {
	return f.submissionDetail, nil
}

func (f *fakeSubmissionReader) ListByUsername(ctx context.Context, params submission.ListByUsernameParams) (submission.SubmissionPage, error) {
	return f.page, nil
}

func TestGetSubmission(t *testing.T) {
	submissionID := uuid.New()
	userID := uuid.New()
	problemID := uuid.New()
	reader := &fakeSubmissionReader{
		submissionDetail: submission.SubmissionDetail{
			Submission: submission.Submission{
				ID:              submissionID,
				UserID:          userID,
				ProblemID:       problemID,
				Language:        "CPP17",
				SourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
				Status:          submission.StatusPending,
				CreatedAt:       time.Date(2026, 5, 15, 1, 2, 3, 0, time.UTC),
			},
			CompileSummary: submission.CompileSummary{
				CompileOutput: "ok",
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/submissions/"+submissionID.String(), nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{SubmissionReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{
		"id":"`+submissionID.String()+`",
		"userId":"`+userID.String()+`",
		"problemId":"`+problemID.String()+`",
		"language":"CPP17",
		"sourceObjectKey":"submissions/2026/05/sub-1/source.cpp.zst",
		"status":"PENDING",
		"createdAt":"2026-05-15T01:02:03Z",
		"compileSummary":{"compileOutput":"ok"},
		"results":[],
		"artifactAvailability":{"sourceObjectKey":"","artifacts":[]}
	}`, rec.Body.String())
}

func TestListUserSubmissions(t *testing.T) {
	reader := &fakeSubmissionReader{
		page: submission.SubmissionPage{
			Items: []submission.Submission{
				{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					ProblemID: uuid.New(),
					Problem: &submission.ProblemSummary{
						ID:    uuid.New(),
						Slug:  "two-sum",
						Title: "Two Sum",
					},
					Language:        "CPP17",
					SourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
					Status:          submission.StatusPending,
					CreatedAt:       time.Date(2026, 5, 15, 1, 2, 3, 0, time.UTC),
				},
			},
			Total:    1,
			Page:     2,
			PageSize: 5,
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/users/demo/submissions?page=2&pageSize=5", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{SubmissionReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"items":[`)
	require.Contains(t, rec.Body.String(), `"page":2`)
	require.Contains(t, rec.Body.String(), `"pageSize":5`)
	require.Contains(t, rec.Body.String(), `"total":1`)
	require.Contains(t, rec.Body.String(), `"problem":{"id":"`)
	require.Contains(t, rec.Body.String(), `"title":"Two Sum"`)
	require.Contains(t, rec.Body.String(), `"language":"CPP17"`)
	require.Contains(t, rec.Body.String(), `"status":"PENDING"`)
}

func TestListUserSubmissionsRejectsBadPageSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/demo/submissions?pageSize=0", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{SubmissionReader: &fakeSubmissionReader{}}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{"error":"invalid pageSize"}`, rec.Body.String())
}
