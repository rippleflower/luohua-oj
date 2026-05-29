package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
)

type Summary struct {
	ID           uuid.UUID
	ProblemNo    int64
	RouteCode    string
	Slug         string
	Title        string
	Difficulty   string
	Tags         []string
	AcceptedRate float64
	UpdatedAt    time.Time
}

type Detail struct {
	ID            uuid.UUID
	ProblemNo     int64
	RouteCode     string
	Slug          string
	Title         string
	Difficulty    string
	StatementJSON json.RawMessage
	SamplesJSON   json.RawMessage
	LimitsJSON    json.RawMessage
	MetadataJSON  json.RawMessage
	UpdatedAt     time.Time
}

type StatementSection struct {
	Kind    string
	Section string
	Content string
}

type Sample struct {
	Input  string
	Output string
	Weight int
}

type AdminProblem struct {
	ID               uuid.UUID
	ProblemNo        int64
	RouteCode        string
	Slug             string
	Title            string
	Difficulty       string
	TimeLimitMs      int
	MemoryLimitKb    int
	Status           string
	CurrentVersionNo int
	IsPublished      bool
	SubmissionCount  int
	AcceptedRate     float64
	UpdatedAt        time.Time
}

type AdminProblemDetail struct {
	AdminProblem
	StatementJSON []StatementSection
	Samples       []Sample
	Tags          []string
}

type CreateAdminInput struct {
	Slug          string
	Title         string
	Difficulty    string
	TimeLimitMs   int
	MemoryLimitKb int
	Reason        string
	IP            string
}

type UpdateAdminInput struct {
	ProblemID     uuid.UUID
	Slug          string
	Title         string
	Difficulty    string
	TimeLimitMs   int
	MemoryLimitKb int
	Reason        string
	IP            string
}

type PublishAdminInput struct {
	ProblemID uuid.UUID
	Reason    string
	IP        string
}

type UpdateAdminContentInput struct {
	ProblemID     uuid.UUID
	StatementJSON []StatementSection
	Samples       []Sample
	Tags          []string
	Reason        string
	IP            string
}

type Reader interface {
	List(ctx context.Context) ([]Summary, error)
	GetBySlug(ctx context.Context, slug string) (Detail, error)
	GetByRouteCode(ctx context.Context, routeCode string) (Detail, error)
}

type AdminManager interface {
	GetAdminDetail(ctx context.Context, problemID uuid.UUID) (AdminProblemDetail, error)
	CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error)
	UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error)
	UpdateAdminContent(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminContentInput) (AdminProblemDetail, error)
	PublishAdmin(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error)
}

type Repository interface {
	ListProblems(ctx context.Context) ([]Summary, error)
	GetProblemBySlug(ctx context.Context, slug string) (Detail, error)
	GetProblemByNumber(ctx context.Context, problemNo int64) (Detail, error)
	GetAdminProblemDetail(ctx context.Context, problemID uuid.UUID) (AdminProblemDetail, error)
	CreateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error)
	UpdateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error)
	UpdateAdminProblemContent(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminContentInput) (AdminProblemDetail, error)
	PublishAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error)
}

type Service struct {
	repo  Repository
	codec RouteCodec
}

func NewService(repo Repository, codec RouteCodec) *Service {
	return &Service{repo: repo, codec: codec}
}

func (s *Service) List(ctx context.Context) ([]Summary, error) {
	items, err := s.repo.ListProblems(ctx)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].RouteCode = s.codec.Encode(items[index].ProblemNo)
	}
	return items, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Detail, error) {
	if strings.TrimSpace(slug) == "" {
		return Detail{}, errors.New("slug is required")
	}
	item, err := s.repo.GetProblemBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return Detail{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) GetByRouteCode(ctx context.Context, routeCode string) (Detail, error) {
	problemNo, err := s.codec.Decode(strings.TrimSpace(routeCode))
	if err != nil {
		return Detail{}, err
	}
	item, err := s.repo.GetProblemByNumber(ctx, problemNo)
	if err != nil {
		return Detail{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) GetAdminDetail(ctx context.Context, problemID uuid.UUID) (AdminProblemDetail, error) {
	if problemID == uuid.Nil {
		return AdminProblemDetail{}, errors.New("problem id is required")
	}
	item, err := s.repo.GetAdminProblemDetail(ctx, problemID)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error) {
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Difficulty = strings.TrimSpace(strings.ToUpper(input.Difficulty))
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminInput(input.Slug, input.Title, input.Difficulty, input.TimeLimitMs, input.MemoryLimitKb); err != nil {
		return AdminProblem{}, err
	}
	item, err := s.repo.CreateAdminProblem(ctx, actor, input)
	if err != nil {
		return AdminProblem{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error) {
	if input.ProblemID == uuid.Nil {
		return AdminProblem{}, errors.New("problem id is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Difficulty = strings.TrimSpace(strings.ToUpper(input.Difficulty))
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminInput(input.Slug, input.Title, input.Difficulty, input.TimeLimitMs, input.MemoryLimitKb); err != nil {
		return AdminProblem{}, err
	}
	item, err := s.repo.UpdateAdminProblem(ctx, actor, input)
	if err != nil {
		return AdminProblem{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) UpdateAdminContent(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminContentInput) (AdminProblemDetail, error) {
	if input.ProblemID == uuid.Nil {
		return AdminProblemDetail{}, errors.New("problem id is required")
	}
	normalizedStatements, err := normalizeStatementSections(input.StatementJSON)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	normalizedSamples, err := normalizeSamples(input.Samples)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	input.StatementJSON = normalizedStatements
	input.Samples = normalizedSamples
	input.Tags = normalizeTags(input.Tags)
	input.Reason = strings.TrimSpace(input.Reason)

	item, err := s.repo.UpdateAdminProblemContent(ctx, actor, input)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func (s *Service) PublishAdmin(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error) {
	if input.ProblemID == uuid.Nil {
		return AdminProblem{}, errors.New("problem id is required")
	}
	input.Reason = strings.TrimSpace(input.Reason)
	item, err := s.repo.PublishAdminProblem(ctx, actor, input)
	if err != nil {
		return AdminProblem{}, err
	}
	item.RouteCode = s.codec.Encode(item.ProblemNo)
	return item, nil
}

func validateAdminInput(slug string, title string, difficulty string, timeLimitMs int, memoryLimitKb int) error {
	if slug == "" {
		return errors.New("slug is required")
	}
	if title == "" {
		return errors.New("title is required")
	}
	switch difficulty {
	case "EASY", "MEDIUM", "HARD":
	default:
		return errors.New("invalid difficulty")
	}
	if timeLimitMs <= 0 {
		return errors.New("timeLimitMs must be positive")
	}
	if memoryLimitKb <= 0 {
		return errors.New("memoryLimitKb must be positive")
	}
	return nil
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeStatementSections(items []StatementSection) ([]StatementSection, error) {
	if len(items) != 4 {
		return nil, errors.New("statementJson must contain four sections")
	}

	required := map[string]struct{}{
		"statement":   {},
		"input":       {},
		"output":      {},
		"constraints": {},
	}
	normalized := make([]StatementSection, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		section := strings.TrimSpace(strings.ToLower(item.Section))
		if _, ok := required[section]; !ok {
			return nil, fmt.Errorf("invalid section: %s", item.Section)
		}
		if _, exists := seen[section]; exists {
			return nil, fmt.Errorf("duplicate section: %s", section)
		}
		kind := strings.TrimSpace(strings.ToLower(item.Kind))
		if kind == "" {
			kind = "markdown"
		}
		if kind != "markdown" {
			return nil, fmt.Errorf("invalid section kind: %s", item.Kind)
		}
		normalized = append(normalized, StatementSection{
			Kind:    kind,
			Section: section,
			Content: strings.TrimSpace(item.Content),
		})
		seen[section] = struct{}{}
	}

	slices.SortFunc(normalized, func(left StatementSection, right StatementSection) int {
		return statementSectionRank(left.Section) - statementSectionRank(right.Section)
	})
	return normalized, nil
}

func normalizeSamples(items []Sample) ([]Sample, error) {
	normalized := make([]Sample, 0, len(items))
	for index, item := range items {
		if item.Weight <= 0 {
			return nil, fmt.Errorf("sample weight must be positive at index %d", index)
		}
		input := strings.ReplaceAll(item.Input, "\r\n", "\n")
		output := strings.ReplaceAll(item.Output, "\r\n", "\n")
		if strings.TrimSpace(input) == "" {
			return nil, fmt.Errorf("sample input is required at index %d", index)
		}
		if strings.TrimSpace(output) == "" {
			return nil, fmt.Errorf("sample output is required at index %d", index)
		}
		normalized = append(normalized, Sample{
			Input:  input,
			Output: output,
			Weight: item.Weight,
		})
	}
	return normalized, nil
}

func normalizeTags(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		tag := strings.ToLower(strings.TrimSpace(item))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return normalized
}

func statementSectionRank(section string) int {
	switch section {
	case "statement":
		return 0
	case "input":
		return 1
	case "output":
		return 2
	case "constraints":
		return 3
	default:
		return 99
	}
}
