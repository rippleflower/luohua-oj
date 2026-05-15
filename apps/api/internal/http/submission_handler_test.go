package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
			ID:           uuid.New(),
			UserID:       userID,
			ProblemID:    problemID,
			Language:     "CPP17",
			SourceObject: "tmp/submissions/source.cpp",
			Status:       submission.StatusPending,
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

	apphttp.NewRouter(apphttp.RouterOptions{SubmissionCreator: creator}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{
		"id":"`+creator.submission.ID.String()+`",
		"userId":"`+userID.String()+`",
		"problemId":"`+problemID.String()+`",
		"language":"CPP17",
		"sourceObject":"tmp/submissions/source.cpp",
		"status":"PENDING"
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

	apphttp.NewRouter(apphttp.RouterOptions{SubmissionCreator: &fakeSubmissionCreator{}}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{"error":"invalid problemId"}`, rec.Body.String())
}

type fakeSubmissionCreator struct {
	submission submission.Submission
}

func (f *fakeSubmissionCreator) Create(ctx context.Context, input submission.CreateInput) (submission.Submission, error) {
	return f.submission, nil
}
