package submission

import (
	"context"
	"fmt"

	db "github.com/example/oj3/apps/api/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SQLRepository struct {
	queries db.Querier
}

func NewSQLRepository(queries db.Querier) *SQLRepository {
	return &SQLRepository{queries: queries}
}

func (r *SQLRepository) CreateSubmission(ctx context.Context, params CreateParams) (Submission, error) {
	created, err := r.queries.CreateSubmission(ctx, db.CreateSubmissionParams{
		ID:           uuidToPG(params.ID),
		UserID:       uuidToPG(params.UserID),
		ProblemID:    uuidToPG(params.ProblemID),
		Language:     db.Language(params.Language),
		SourceObject: params.SourceObject,
	})
	if err != nil {
		return Submission{}, err
	}

	return submissionFromDB(created)
}

func submissionFromDB(model db.Submission) (Submission, error) {
	id, err := uuidFromPG(model.ID)
	if err != nil {
		return Submission{}, fmt.Errorf("submission id: %w", err)
	}
	userID, err := uuidFromPG(model.UserID)
	if err != nil {
		return Submission{}, fmt.Errorf("user id: %w", err)
	}
	problemID, err := uuidFromPG(model.ProblemID)
	if err != nil {
		return Submission{}, fmt.Errorf("problem id: %w", err)
	}

	return Submission{
		ID:           id,
		UserID:       userID,
		ProblemID:    problemID,
		Language:     string(model.Language),
		SourceObject: model.SourceObject,
		Status:       Status(model.Status),
	}, nil
}

func uuidToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPG(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, fmt.Errorf("uuid is null")
	}
	return uuid.UUID(id.Bytes), nil
}
