package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}
