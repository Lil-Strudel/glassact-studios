package modules

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/config"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testContext struct {
	app       *app.Application
	handler   http.Handler
	db        data.Models
	container testcontainers.Container
}

func setupTestApp(t *testing.T) (*testContext, func()) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgis/postgis:18-3.6",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	for i := 0; i < 30; i++ {
		err = pool.Ping(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	stdb := stdlib.OpenDB(*cfg.ConnConfig)

	err = runMigrations(stdb)
	require.NoError(t, err)

	db := data.NewModels(pool, stdb)

	testApp := &app.Application{
		Cfg: &config.Config{
			Env:        "test",
			Port:       8080,
			BaseURL:    "http://localhost:3000",
			AuthSecret: "test-secret-key-at-least-32-characters-long",
		},
		Db:       db,
		Err:      app.AppError,
		Log:      slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
		Wg:       sync.WaitGroup{},
		S3:       nil,
	}

	cleanup := func() {
		pool.Close()
		stdb.Close()
		container.Terminate(ctx) //nolint:errcheck
	}

	return &testContext{
		app:       testApp,
		handler:   GetRoutes(testApp),
		db:        db,
		container: container,
	}, cleanup
}

func runMigrations(db *sql.DB) error {
	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationPath, err := filepath.Abs("../../../libs/data/migrations")
	if err != nil {
		return fmt.Errorf("failed to resolve migrations directory: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

type testRequest struct {
	method  string
	path    string
	body    interface{}
	token   string
	headers map[string]string
}

type testResponse struct {
	statusCode int
	body       []byte
	parsed     interface{}
}

func (tc *testContext) request(req testRequest) *testResponse {
	var bodyReader io.Reader
	if req.body != nil {
		bodyBytes, _ := json.Marshal(req.body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq := httptest.NewRequest(req.method, req.path, bodyReader)
	httpReq.Header.Set("Content-Type", "application/json")

	if req.token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", req.token))
	}

	for k, v := range req.headers {
		httpReq.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	tc.handler.ServeHTTP(w, httpReq)

	respBody := w.Body.Bytes()
	var parsed interface{}
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &parsed)
	}

	return &testResponse{
		statusCode: w.Code,
		body:       respBody,
		parsed:     parsed,
	}
}

func seedTestData(t *testing.T, ctx *testContext) (*data.DealershipUser, string) {
	dealership := &data.Dealership{
		Name: "Test Dealership",
		Address: data.Address{
			Street:     "123 Main St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	err := ctx.db.Dealerships.Insert(dealership)
	require.NoError(t, err)

	user := &data.DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		Role:         "admin",
		IsActive:     true,
	}
	err = ctx.db.DealershipUsers.Insert(user)
	require.NoError(t, err)

	token, err := ctx.db.DealershipTokens.New(user.ID, 2*time.Hour, data.DealershipScopeAccess)
	require.NoError(t, err)

	return user, token.Plaintext
}

func TestAPIEndpoints(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	user, accessToken := seedTestData(t, testCtx)

	t.Run("Auth Module", func(t *testing.T) {
		t.Run("GET /api/auth/google", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/auth/google",
			})

			assert.Equal(t, http.StatusFound, resp.statusCode)
			t.Logf("✓ GET /api/auth/google (%d)", resp.statusCode)
		})

		t.Run("GET /api/auth/microsoft", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/auth/microsoft",
			})

			assert.Equal(t, http.StatusFound, resp.statusCode)
			t.Logf("✓ GET /api/auth/microsoft (%d)", resp.statusCode)
		})

		t.Run("POST /api/auth/magic-link (invalid email)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/auth/magic-link",
				body: map[string]string{
					"email": "nonexistent@example.com",
				},
			})

			assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
			t.Logf("✓ POST /api/auth/magic-link (invalid email) (%d)", resp.statusCode)
		})

		t.Run("POST /api/auth/token/access", func(t *testing.T) {
			refreshToken, err := testCtx.db.DealershipTokens.New(user.ID, 24*time.Hour, data.DealershipScopeRefresh)
			require.NoError(t, err)

			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/auth/token/access",
				headers: map[string]string{
					"Cookie": fmt.Sprintf("refresh_token=%s", refreshToken.Plaintext),
				},
			})

			var body map[string]any
			_ = json.Unmarshal(resp.body, &body)

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			assert.NotNil(t, body["access_token"])
			assert.NotNil(t, body["access_token_exp"])
		})

		t.Run("GET /api/auth/logout", func(t *testing.T) {
			refreshToken, err := testCtx.db.DealershipTokens.New(user.ID, 24*time.Hour, data.DealershipScopeRefresh)
			require.NoError(t, err)

			req := httptest.NewRequest("GET", "/api/auth/logout", nil)
			req.Header.Set("Cookie", fmt.Sprintf("refresh_token=%s", refreshToken.Plaintext))

			w := httptest.NewRecorder()
			testCtx.handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusFound, w.Code)
			t.Logf("✓ GET /api/auth/logout (%d)", w.Code)
		})
	})

	t.Run("Dealership Module", func(t *testing.T) {
		t.Run("GET /api/dealership", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/dealership",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/dealership (%d)", resp.statusCode)

			var dealerships []map[string]interface{}
			_ = json.Unmarshal(resp.body, &dealerships)
			assert.Greater(t, len(dealerships), 0)
		})

		t.Run("POST /api/dealership", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/dealership",
				body: map[string]interface{}{
					"name": "New Test Dealership",
					"address": map[string]interface{}{
						"street":      "456 Oak Ave",
						"city":        "New City",
						"state":       "NC",
						"postal_code": "54321",
						"country":     "USA",
						"latitude":    40.0,
						"longitude":   -75.0,
					},
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/dealership (%d)", resp.statusCode)

			var dealership map[string]interface{}
			_ = json.Unmarshal(resp.body, &dealership)
			assert.NotNil(t, dealership["uuid"])
		})

		dealerships, _ := testCtx.db.Dealerships.GetAll()
		if len(dealerships) > 0 {
			t.Run("GET /api/dealership/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/dealership/%s", dealerships[0].UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ GET /api/dealership/{uuid} (%d)", resp.statusCode)
			})
		}
	})

	t.Run("Inlay Module", func(t *testing.T) {
		dealerships, _ := testCtx.db.Dealerships.GetAll()
		require.Greater(t, len(dealerships), 0)

		project := &data.Project{
			Name:         "Test Project",
			Status:       data.ProjectStatuses.Draft,
			DealershipID: dealerships[0].ID,
		}
		_ = testCtx.db.Projects.Insert(project)

		t.Run("GET /api/inlay", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/inlay",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/inlay (%d)", resp.statusCode)
		})

		t.Run("POST /api/inlay", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/inlay",
				body: map[string]interface{}{
					"project_id":  project.ID,
					"name":        "Test Inlay",
					"type":        "catalog",
					"preview_url": "https://example.com/preview.png",
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/inlay (%d)", resp.statusCode)
		})

		inlays, _ := testCtx.db.Inlays.GetAll()
		if len(inlays) > 0 {
			t.Run("GET /api/inlay/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/inlay/%s", inlays[0].UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ GET /api/inlay/{uuid} (%d)", resp.statusCode)
			})
		}
	})

	t.Run("Inlay-Chat Module", func(t *testing.T) {
		projects, _ := testCtx.db.Projects.GetAll()
		require.Greater(t, len(projects), 0)

		inlays, _ := testCtx.db.Inlays.GetAll()
		require.Greater(t, len(inlays), 0)

		t.Run("GET /api/inlay-chat", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/inlay-chat",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/inlay-chat (%d)", resp.statusCode)
		})

		t.Run("GET /api/inlay-chat/inlay/{uuid}", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/inlay-chat/inlay/%s", inlays[0].UUID),
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/inlay-chat/inlay/{uuid} (%d)", resp.statusCode)
		})

		t.Run("POST /api/inlay-chat", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/inlay-chat",
				body: map[string]interface{}{
					"inlay_id":     inlays[0].ID,
					"message_type": "system",
					"message":      "Test message",
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/inlay-chat (%d)", resp.statusCode)
		})

		chats, _ := testCtx.db.InlayChats.GetAll()
		if len(chats) > 0 {
			t.Run("GET /api/inlay-chat/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/inlay-chat/%s", chats[0].UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ GET /api/inlay-chat/{uuid} (%d)", resp.statusCode)
			})
		}
	})

	t.Run("Project Module", func(t *testing.T) {
		dealerships, _ := testCtx.db.Dealerships.GetAll()
		require.Greater(t, len(dealerships), 0)

		t.Run("GET /api/project", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/project",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/project (%d)", resp.statusCode)
		})

		t.Run("POST /api/project", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/project",
				body: map[string]interface{}{
					"name":          "New Test Project",
					"dealership_id": dealerships[0].ID,
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/project (%d)", resp.statusCode)
		})

		t.Run("POST /api/project/with-inlays", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/project/with-inlays",
				body: map[string]interface{}{
					"name":          "Project with Inlays",
					"dealership_id": dealerships[0].ID,
					"inlays": []map[string]interface{}{
						{
							"name":        "Inlay 1",
							"type":        "catalog",
							"preview_url": "https://example.com/preview.png",
						},
					},
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/project/with-inlays (%d)", resp.statusCode)
		})

		projects, _ := testCtx.db.Projects.GetAll()
		if len(projects) > 0 {
			t.Run("GET /api/project/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/project/%s", projects[0].UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ GET /api/project/{uuid} (%d)", resp.statusCode)
			})
		}
	})

	t.Run("User Module", func(t *testing.T) {
		t.Run("GET /api/user/self", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/user/self",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/user/self (%d)", resp.statusCode)
		})

		t.Run("GET /api/dealership-user", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/dealership-user",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/dealership-user (%d)", resp.statusCode)
		})

		t.Run("POST /api/dealership-user", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/dealership-user",
				body: map[string]interface{}{
					"name":  "New Dealership User",
					"email": fmt.Sprintf("newuser%d@example.com", time.Now().UnixNano()),
					"role":  "user",
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/dealership-user (%d)", resp.statusCode)
		})

		dealershipUsers, _ := testCtx.db.DealershipUsers.GetAll()
		if len(dealershipUsers) > 1 {
			targetUser := dealershipUsers[1]
			t.Run("GET /api/dealership-user/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/dealership-user/%s", targetUser.UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ GET /api/dealership-user/{uuid} (%d)", resp.statusCode)
			})

			t.Run("PATCH /api/dealership-user/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "PATCH",
					path:   fmt.Sprintf("/api/dealership-user/%s", targetUser.UUID),
					body: map[string]interface{}{
						"name": "Updated User Name",
					},
					token: accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ PATCH /api/dealership-user/{uuid} (%d)", resp.statusCode)
			})

			t.Run("DELETE /api/dealership-user/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "DELETE",
					path:   fmt.Sprintf("/api/dealership-user/%s", targetUser.UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusNoContent, resp.statusCode)
				t.Logf("✓ DELETE /api/dealership-user/{uuid} (%d)", resp.statusCode)
			})
		}

		t.Run("POST /api/internal-user", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/internal-user",
				body: map[string]interface{}{
					"name":  "New Internal User",
					"email": fmt.Sprintf("newinternal%d@example.com", time.Now().UnixNano()),
					"role":  "user",
				},
				token: accessToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/internal-user (%d)", resp.statusCode)
		})

		internalUsers, _ := testCtx.db.InternalUsers.GetAll()
		if len(internalUsers) > 0 {
			targetUser := internalUsers[0]
			t.Run("PATCH /api/internal-user/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "PATCH",
					path:   fmt.Sprintf("/api/internal-user/%s", targetUser.UUID),
					body: map[string]interface{}{
						"name": "Updated Internal User",
					},
					token: accessToken,
				})

				assert.Equal(t, http.StatusOK, resp.statusCode)
				t.Logf("✓ PATCH /api/internal-user/{uuid} (%d)", resp.statusCode)
			})

			t.Run("DELETE /api/internal-user/{uuid}", func(t *testing.T) {
				resp := testCtx.request(testRequest{
					method: "DELETE",
					path:   fmt.Sprintf("/api/internal-user/%s", targetUser.UUID),
					token:  accessToken,
				})

				assert.Equal(t, http.StatusNoContent, resp.statusCode)
				t.Logf("✓ DELETE /api/internal-user/{uuid} (%d)", resp.statusCode)
			})
		}
	})

	t.Run("Upload Module", func(t *testing.T) {
		t.Run("POST /api/upload (missing file)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/upload",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/upload (missing file) (%d)", resp.statusCode)
		})
	})

	t.Run("Error Handling", func(t *testing.T) {
		t.Run("Protected endpoint without token", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/dealership",
			})

			assert.Equal(t, http.StatusUnauthorized, resp.statusCode)
			t.Logf("✓ Protected endpoint without token (%d)", resp.statusCode)
		})

		t.Run("Nonexistent route", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/nonexistent",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ Nonexistent route (%d)", resp.statusCode)
		})
	})
}
