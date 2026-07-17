//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	catalogpostgres "github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/postgres"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/migration"
	"github.com/WeiGuang-2099/LaunchRail/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestServiceHTTPRoutesPersistLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pool.Ping() error = %v", err)
	}
	if err := migration.New(pool, migrations.Files).Up(ctx); err != nil {
		t.Fatalf("migration Up() error = %v", err)
	}

	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() error = %v", err)
	}
	t.Cleanup(func() {
		if err := transaction.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("transaction.Rollback() error = %v", err)
		}
	})

	repository := catalogpostgres.NewServiceRepository(transaction)
	now := time.Date(2026, time.July, 17, 5, 0, 0, 0, time.UTC)
	catalog := application.NewCatalog(repository, func() time.Time {
		now = now.Add(time.Second)
		return now
	})
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := NewRouter(logger, pool.Ping, catalog)

	create := integrationRequest(t, router, http.MethodPost, "/api/v1/services", `{
		"serviceKey":"integration-api",
		"displayName":"Integration API",
		"description":"Created through HTTP",
		"ownerTeam":"platform-team",
		"repositoryUrl":"https://github.com/example/integration-api",
		"lifecycle":"development"
	}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, body = %s", create.Code, create.Body.String())
	}
	var created serviceResponse
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created service: %v", err)
	}

	list := integrationRequest(t, router, http.MethodGet, "/api/v1/services?status=active&limit=20&offset=0", "")
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(created.ID.String())) {
		t.Fatalf("GET list status = %d, body = %s", list.Code, list.Body.String())
	}

	update := integrationRequest(t, router, http.MethodPut, "/api/v1/services/"+created.ID.String(), `{
		"displayName":"Integration API v2",
		"description":"Updated through HTTP",
		"ownerTeam":"runtime-team",
		"repositoryUrl":"https://github.com/example/integration-api",
		"lifecycle":"production"
	}`)
	if update.Code != http.StatusOK || !bytes.Contains(update.Body.Bytes(), []byte("Integration API v2")) {
		t.Fatalf("PUT status = %d, body = %s", update.Code, update.Body.String())
	}

	archive := integrationRequest(t, router, http.MethodDelete, "/api/v1/services/"+created.ID.String(), "")
	if archive.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, body = %s", archive.Code, archive.Body.String())
	}

	archivedList := integrationRequest(t, router, http.MethodGet, "/api/v1/services?status=archived&limit=20&offset=0", "")
	if archivedList.Code != http.StatusOK || !bytes.Contains(archivedList.Body.Bytes(), []byte(created.ID.String())) {
		t.Fatalf("GET archived list status = %d, body = %s", archivedList.Code, archivedList.Body.String())
	}
}

func integrationRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
