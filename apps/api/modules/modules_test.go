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

func seedTestData(t *testing.T, ctx *testContext) (*data.DealershipUser, string, *data.InternalUser, string) {
	dealership := &data.Dealership{
		Name: "Test Dealership",
		Address: data.Address{
			Street:     "123 Main St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "US",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	err := ctx.db.Dealerships.Insert(dealership)
	require.NoError(t, err)

	dealershipUser := &data.DealershipUser{
		DealershipID: dealership.ID,
		Name:         "Test User",
		Email:        fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		Role:         "admin",
		IsActive:     true,
	}
	err = ctx.db.DealershipUsers.Insert(dealershipUser)
	require.NoError(t, err)

	dealershipToken, err := ctx.db.DealershipTokens.New(dealershipUser.ID, 2*time.Hour, data.DealershipScopeAccess)
	require.NoError(t, err)

	// Create internal admin user
	internalUser := &data.InternalUser{
		Name:     "Internal Admin",
		Email:    fmt.Sprintf("admin%d@example.com", time.Now().UnixNano()),
		Role:     "admin",
		IsActive: true,
	}
	err = ctx.db.InternalUsers.Insert(internalUser)
	require.NoError(t, err)

	internalToken, err := ctx.db.InternalTokens.New(internalUser.ID, 2*time.Hour, data.InternalScopeAccess)
	require.NoError(t, err)

	return dealershipUser, dealershipToken.Plaintext, internalUser, internalToken.Plaintext
}

func seedPriceGroup(t *testing.T, ctx *testContext, name string) *data.PriceGroup {
	priceGroup := &data.PriceGroup{
		Name:           name,
		BasePriceCents: 10000, // $100.00
		IsActive:       true,
	}
	err := ctx.db.PriceGroups.Insert(priceGroup)
	require.NoError(t, err)
	return priceGroup
}

func seedCatalogItem(t *testing.T, ctx *testContext, priceGroupID int, catalogCode string) *data.CatalogItem {
	item := &data.CatalogItem{
		CatalogCode:         catalogCode,
		Name:                "Test Item " + catalogCode,
		Category:            "test",
		DefaultWidth:        10.0,
		DefaultHeight:       10.0,
		MinWidth:            5.0,
		MinHeight:           5.0,
		DefaultPriceGroupID: priceGroupID,
		SvgURL:              "https://example.com/test.svg",
		IsActive:            true,
	}
	err := ctx.db.CatalogItems.Insert(item)
	require.NoError(t, err)
	return item
}

func TestAPIEndpoints(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, internalUser, internalAdminToken := seedTestData(t, testCtx)
	_ = dealershipUser // Keep for compatibility with existing tests
	_ = internalUser   // Keep for compatibility
	accessToken := dealershipToken
	user := dealershipUser

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
						"country":     "US",
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
					"name": "New Test Project",
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
					"name": "Project with Inlays",
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

	t.Run("Catalog Module", func(t *testing.T) {
		// Create test data
		priceGroup := seedPriceGroup(t, testCtx, "Test Price Group")
		catalogItem := seedCatalogItem(t, testCtx, priceGroup.ID, "TEST-001")

		t.Run("POST /api/catalog (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/catalog",
				body: map[string]interface{}{
					"catalog_code":           "NEW-ITEM-001",
					"name":                   "New Catalog Item",
					"category":               "stained-glass",
					"default_width":          12.5,
					"default_height":         15.0,
					"min_width":              10.0,
					"min_height":             12.0,
					"default_price_group_id": priceGroup.ID,
					"svg_url":                "https://example.com/new.svg",
					"is_active":              true,
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/catalog (happy path) (%d)", resp.statusCode)

			var item map[string]interface{}
			_ = json.Unmarshal(resp.body, &item)
			assert.Equal(t, "NEW-ITEM-001", item["catalog_code"])
		})

		t.Run("POST /api/catalog (missing required field)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/catalog",
				body: map[string]interface{}{
					"catalog_code": "INVALID-001",
					// missing name
					"category":               "stained-glass",
					"default_width":          12.5,
					"default_height":         15.0,
					"min_width":              10.0,
					"min_height":             12.0,
					"default_price_group_id": priceGroup.ID,
					"svg_url":                "https://example.com/new.svg",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/catalog (missing required field) (%d)", resp.statusCode)
		})

		t.Run("POST /api/catalog (invalid dimensions)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/catalog",
				body: map[string]interface{}{
					"catalog_code":           "DIM-TEST-001",
					"name":                   "Dimension Test",
					"category":               "stained-glass",
					"default_width":          5.0, // Less than min
					"default_height":         8.0, // Less than min
					"min_width":              10.0,
					"min_height":             12.0,
					"default_price_group_id": priceGroup.ID,
					"svg_url":                "https://example.com/new.svg",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/catalog (invalid dimensions) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog (happy path) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.NotNil(t, body["items"])
			assert.NotNil(t, body["total"])
		})

		t.Run("GET /api/catalog (with pagination)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog?limit=10&offset=0",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog (with pagination) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.Equal(t, float64(10), body["limit"])
			assert.Equal(t, float64(0), body["offset"])
		})

		t.Run("GET /api/catalog (with search filter)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog?search=TEST",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog (with search filter) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog/{uuid} (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/catalog/%s", catalogItem.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog/{uuid} (happy path) (%d)", resp.statusCode)

			var item map[string]interface{}
			_ = json.Unmarshal(resp.body, &item)
			assert.Equal(t, catalogItem.UUID, item["uuid"])
		})

		t.Run("GET /api/catalog/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog/00000000-0000-0000-0000-000000000000",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ GET /api/catalog/{uuid} (not found) (%d)", resp.statusCode)
		})

		t.Run("PATCH /api/catalog/{uuid} (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "PATCH",
				path:   fmt.Sprintf("/api/catalog/%s", catalogItem.UUID),
				body: map[string]interface{}{
					"name": "Updated Catalog Item",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ PATCH /api/catalog/{uuid} (happy path) (%d)", resp.statusCode)

			var item map[string]interface{}
			_ = json.Unmarshal(resp.body, &item)
			assert.Equal(t, "Updated Catalog Item", item["name"])
		})

		t.Run("PATCH /api/catalog/{uuid} (invalid dimensions)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "PATCH",
				path:   fmt.Sprintf("/api/catalog/%s", catalogItem.UUID),
				body: map[string]interface{}{
					"default_width":  3.0, // Less than min
					"default_height": 3.0, // Less than min
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ PATCH /api/catalog/{uuid} (invalid dimensions) (%d)", resp.statusCode)
		})

		t.Run("PATCH /api/catalog/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "PATCH",
				path:   "/api/catalog/00000000-0000-0000-0000-000000000000",
				body: map[string]interface{}{
					"name": "Updated",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ PATCH /api/catalog/{uuid} (not found) (%d)", resp.statusCode)
		})

		t.Run("DELETE /api/catalog/{uuid} (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "DELETE",
				path:   fmt.Sprintf("/api/catalog/%s", catalogItem.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ DELETE /api/catalog/{uuid} (happy path) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.Equal(t, true, body["success"])
		})

		t.Run("DELETE /api/catalog/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "DELETE",
				path:   "/api/catalog/00000000-0000-0000-0000-000000000000",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ DELETE /api/catalog/{uuid} (not found) (%d)", resp.statusCode)
		})

		t.Run("POST /api/catalog/{uuid}/tags (happy path)", func(t *testing.T) {
			catalogItem2 := seedCatalogItem(t, testCtx, priceGroup.ID, "TAG-TEST-001")
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem2.UUID),
				body: map[string]interface{}{
					"tag": "premium",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ POST /api/catalog/{uuid}/tags (happy path) (%d)", resp.statusCode)

			var tags []interface{}
			_ = json.Unmarshal(resp.body, &tags)
			assert.Greater(t, len(tags), 0)
		})

		t.Run("POST /api/catalog/{uuid}/tags (duplicate tag error)", func(t *testing.T) {
			catalogItem3 := seedCatalogItem(t, testCtx, priceGroup.ID, "DUPTEST-001")
			// Add tag first time
			testCtx.request(testRequest{
				method: "POST",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem3.UUID),
				body: map[string]interface{}{
					"tag": "duplicate-tag",
				},
				token: internalAdminToken,
			})

			// Try to add same tag again
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem3.UUID),
				body: map[string]interface{}{
					"tag": "duplicate-tag",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/catalog/{uuid}/tags (duplicate tag error) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog/{uuid}/tags (happy path)", func(t *testing.T) {
			catalogItem4 := seedCatalogItem(t, testCtx, priceGroup.ID, "GETTAG-001")
			testCtx.request(testRequest{
				method: "POST",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem4.UUID),
				body: map[string]interface{}{
					"tag": "test-tag",
				},
				token: internalAdminToken,
			})

			resp := testCtx.request(testRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem4.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog/{uuid}/tags (happy path) (%d)", resp.statusCode)
		})

		t.Run("DELETE /api/catalog/{uuid}/tags/{tag} (happy path)", func(t *testing.T) {
			catalogItem5 := seedCatalogItem(t, testCtx, priceGroup.ID, "DELTAG-001")
			testCtx.request(testRequest{
				method: "POST",
				path:   fmt.Sprintf("/api/catalog/%s/tags", catalogItem5.UUID),
				body: map[string]interface{}{
					"tag": "remove-me",
				},
				token: internalAdminToken,
			})

			resp := testCtx.request(testRequest{
				method: "DELETE",
				path:   fmt.Sprintf("/api/catalog/%s/tags/remove-me", catalogItem5.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ DELETE /api/catalog/{uuid}/tags/{tag} (happy path) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog/browse (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog/browse",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog/browse (happy path) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog/categories (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog/categories",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog/categories (happy path) (%d)", resp.statusCode)
		})

		t.Run("GET /api/catalog/tags (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/catalog/tags",
				token:  accessToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/catalog/tags (happy path) (%d)", resp.statusCode)
		})
	})

	t.Run("Price Group Module", func(t *testing.T) {
		t.Run("POST /api/price-groups (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/price-groups",
				body: map[string]interface{}{
					"name":             "Premium Tier",
					"base_price_cents": 25000, // $250.00
					"description":      "Premium pricing tier",
					"is_active":        true,
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusCreated, resp.statusCode)
			t.Logf("✓ POST /api/price-groups (happy path) (%d)", resp.statusCode)

			var priceGroup map[string]interface{}
			_ = json.Unmarshal(resp.body, &priceGroup)
			assert.Equal(t, "Premium Tier", priceGroup["name"])
			assert.Equal(t, float64(25000), priceGroup["base_price_cents"])
		})

		t.Run("POST /api/price-groups (missing required field)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/price-groups",
				body: map[string]interface{}{
					// missing name
					"base_price_cents": 15000,
					"is_active":        true,
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/price-groups (missing required field) (%d)", resp.statusCode)
		})

		t.Run("POST /api/price-groups (invalid price)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "POST",
				path:   "/api/price-groups",
				body: map[string]interface{}{
					"name":             "Bad Price",
					"base_price_cents": 0, // Must be > 0
					"is_active":        true,
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusBadRequest, resp.statusCode)
			t.Logf("✓ POST /api/price-groups (invalid price) (%d)", resp.statusCode)
		})

		priceGroup := seedPriceGroup(t, testCtx, "Standard Tier")

		t.Run("GET /api/price-groups (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/price-groups",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/price-groups (happy path) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.NotNil(t, body["items"])
			assert.NotNil(t, body["total"])
		})

		t.Run("GET /api/price-groups (with pagination)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/price-groups?limit=10&offset=0",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/price-groups (with pagination) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.Equal(t, float64(10), body["limit"])
			assert.Equal(t, float64(0), body["offset"])
		})

		t.Run("GET /api/price-groups/{uuid} (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/price-groups/%s", priceGroup.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ GET /api/price-groups/{uuid} (happy path) (%d)", resp.statusCode)

			var pg map[string]interface{}
			_ = json.Unmarshal(resp.body, &pg)
			assert.Equal(t, priceGroup.UUID, pg["uuid"])
		})

		t.Run("GET /api/price-groups/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "GET",
				path:   "/api/price-groups/00000000-0000-0000-0000-000000000000",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ GET /api/price-groups/{uuid} (not found) (%d)", resp.statusCode)
		})

		t.Run("PATCH /api/price-groups/{uuid} (happy path)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "PATCH",
				path:   fmt.Sprintf("/api/price-groups/%s", priceGroup.UUID),
				body: map[string]interface{}{
					"name": "Updated Standard Tier",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ PATCH /api/price-groups/{uuid} (happy path) (%d)", resp.statusCode)

			var pg map[string]interface{}
			_ = json.Unmarshal(resp.body, &pg)
			assert.Equal(t, "Updated Standard Tier", pg["name"])
		})

		t.Run("PATCH /api/price-groups/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "PATCH",
				path:   "/api/price-groups/00000000-0000-0000-0000-000000000000",
				body: map[string]interface{}{
					"name": "Updated",
				},
				token: internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ PATCH /api/price-groups/{uuid} (not found) (%d)", resp.statusCode)
		})

		t.Run("DELETE /api/price-groups/{uuid} (happy path)", func(t *testing.T) {
			priceGroupToDelete := seedPriceGroup(t, testCtx, "To Delete")

			resp := testCtx.request(testRequest{
				method: "DELETE",
				path:   fmt.Sprintf("/api/price-groups/%s", priceGroupToDelete.UUID),
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusOK, resp.statusCode)
			t.Logf("✓ DELETE /api/price-groups/{uuid} (happy path) (%d)", resp.statusCode)

			var body map[string]interface{}
			_ = json.Unmarshal(resp.body, &body)
			assert.Equal(t, true, body["success"])
		})

		t.Run("DELETE /api/price-groups/{uuid} (not found)", func(t *testing.T) {
			resp := testCtx.request(testRequest{
				method: "DELETE",
				path:   "/api/price-groups/00000000-0000-0000-0000-000000000000",
				token:  internalAdminToken,
			})

			assert.Equal(t, http.StatusNotFound, resp.statusCode)
			t.Logf("✓ DELETE /api/price-groups/{uuid} (not found) (%d)", resp.statusCode)
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
