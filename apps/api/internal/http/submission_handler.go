package http

import (
	"encoding/json"
	"net/http"

	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
)

type createSubmissionRequest struct {
	UserID    string `json:"userId"`
	ProblemID string `json:"problemId"`
	Language  string `json:"language"`
	Source    string `json:"source"`
}

type createSubmissionResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	ProblemID    string `json:"problemId"`
	Language     string `json:"language"`
	SourceObject string `json:"sourceObject"`
	Status       string `json:"status"`
}

func createSubmissionHandler(creator submission.Creator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createSubmissionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		userID, err := uuid.Parse(request.UserID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid userId")
			return
		}
		problemID, err := uuid.Parse(request.ProblemID)
		if err != nil {
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
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, createSubmissionResponse{
			ID:           created.ID.String(),
			UserID:       created.UserID.String(),
			ProblemID:    created.ProblemID.String(),
			Language:     created.Language,
			SourceObject: created.SourceObject,
			Status:       string(created.Status),
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
