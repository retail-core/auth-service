package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/logger"
)

func TestPingEndpoint(t *testing.T) {
	config.LoadConfig()
	// router := api.NewRouter()
	logger.Init("testing")

	// req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
