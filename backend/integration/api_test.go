//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"taskflow/internal/config"
	"taskflow/internal/db"
	"taskflow/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	sharedPool *pgxpool.Pool
	sharedCfg  *config.Config
)

func TestMain(m *testing.M) {
	sharedCfg = buildTestConfig()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	var err error
	sharedPool, err = db.ConnectPool(context.Background(), sharedCfg, log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration: db connect:", err)
		os.Exit(1)
	}

	code := m.Run()
	sharedPool.Close()
	os.Exit(code)
}

func buildTestConfig() *config.Config {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://taskflow:taskflow@localhost:5432/taskflow_test?sslmode=disable"
	}
	d := func(dur time.Duration) config.Duration { return config.Duration{Duration: dur} }
	return &config.Config{
		App: config.AppConfig{Name: "taskflow-integration", Env: "test"},
		Server: config.ServerConfig{
			Port: 0, ReadTimeout: d(5 * time.Second),
			WriteTimeout: d(5 * time.Second), IdleTimeout: d(10 * time.Second),
			ShutdownTimeout: d(5 * time.Second),
		},
		DB: config.DBConfig{
			URL: dbURL,
			Pool: config.PoolConfig{
				MaxConns: 5, MinConns: 1,
				MaxConnLifetime:   d(5 * time.Minute),
				MaxConnIdleTime:   d(time.Minute),
				HealthCheckPeriod: d(30 * time.Second),
				ConnectTimeout:    d(5 * time.Second),
				StatementTimeout:  d(10 * time.Second),
			},
		},
		Auth: config.AuthConfig{
			JWTSecret: "integration-test-secret-do-not-use-in-production",
			TokenTTL:  d(time.Hour),
		},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
		Seed: config.SeedConfig{},
	}
}

type testEnv struct {
	srv *httptest.Server
}

func newEnv(t *testing.T) *testEnv {
	t.Helper()
	truncate(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(server.New(sharedCfg, sharedPool, log).Handler())
	t.Cleanup(func() {
		srv.Close()
		truncate(t)
	})
	return &testEnv{srv: srv}
}

func truncate(t *testing.T) {
	t.Helper()
	_, err := sharedPool.Exec(context.Background(),
		`TRUNCATE users, projects, tasks, user_sessions CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func (e *testEnv) do(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var rb io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rb = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.srv.URL+path, rb)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func mustDecode(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode == want {
		return
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	t.Fatalf("want HTTP %d got %d — body: %s", want, resp.StatusCode, b)
}

type authBody struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

func (e *testEnv) register(t *testing.T, name, email, password string) authBody {
	t.Helper()
	resp := e.do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"name": name, "email": email, "password": password,
	})
	assertStatus(t, resp, http.StatusCreated)
	var a authBody
	mustDecode(t, resp, &a)
	return a
}

func (e *testEnv) loginRaw(t *testing.T, email, password string) *http.Response {
	t.Helper()
	return e.do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email": email, "password": password,
	})
}

func (e *testEnv) createProject(t *testing.T, token, name string) map[string]any {
	t.Helper()
	resp := e.do(t, http.MethodPost, "/api/projects", token, map[string]string{"name": name})
	assertStatus(t, resp, http.StatusCreated)
	var p map[string]any
	mustDecode(t, resp, &p)
	return p
}

func (e *testEnv) createTask(t *testing.T, token, projectID, title, status string) map[string]any {
	t.Helper()
	body := map[string]string{"title": title}
	if status != "" {
		body["status"] = status
	}
	resp := e.do(t, http.MethodPost, "/api/projects/"+projectID+"/tasks", token, body)
	assertStatus(t, resp, http.StatusCreated)
	var task map[string]any
	mustDecode(t, resp, &task)
	return task
}

func TestAuth_Register(t *testing.T) {
	e := newEnv(t)

	t.Run("success", func(t *testing.T) {
		resp := e.do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
			"name": "Alice", "email": "alice@example.com", "password": "password123",
		})
		assertStatus(t, resp, http.StatusCreated)
		var body authBody
		mustDecode(t, resp, &body)
		if body.Token == "" {
			t.Fatal("token must be non-empty")
		}
		if body.User.Email != "alice@example.com" {
			t.Fatalf("user.email: got %q want %q", body.User.Email, "alice@example.com")
		}
		if body.User.ID == "" {
			t.Fatal("user.id must be non-empty")
		}
	})

	t.Run("duplicate email returns 400 with field error", func(t *testing.T) {
		e.register(t, "Alice", "dup@example.com", "password123")

		resp := e.do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
			"name": "Alice2", "email": "dup@example.com", "password": "password456",
		})
		assertStatus(t, resp, http.StatusBadRequest)
		var body map[string]any
		mustDecode(t, resp, &body)
		if body["error"] != "validation failed" {
			t.Fatalf("error: got %q", body["error"])
		}
		fields, _ := body["fields"].(map[string]any)
		if fields["email"] == nil {
			t.Fatalf("expected fields.email, got %v", fields)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		cases := []struct {
			desc  string
			body  map[string]string
			field string
		}{
			{"missing name", map[string]string{"email": "a@b.com", "password": "12345678"}, "name"},
			{"missing email", map[string]string{"name": "Bob", "password": "12345678"}, "email"},
			{"invalid email format", map[string]string{"name": "Bob", "email": "not-an-email", "password": "12345678"}, "email"},
			{"password too short", map[string]string{"name": "Bob", "email": "b@b.com", "password": "short"}, "password"},
		}
		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				resp := e.do(t, http.MethodPost, "/api/auth/register", "", tc.body)
				assertStatus(t, resp, http.StatusBadRequest)
				var body map[string]any
				mustDecode(t, resp, &body)
				fields, _ := body["fields"].(map[string]any)
				if fields[tc.field] == nil {
					t.Fatalf("expected fields.%s, got %v", tc.field, fields)
				}
			})
		}
	})
}

func TestAuth_Login(t *testing.T) {
	e := newEnv(t)
	e.register(t, "Alice", "alice@example.com", "password123")

	t.Run("success returns token and user", func(t *testing.T) {
		resp := e.loginRaw(t, "alice@example.com", "password123")
		assertStatus(t, resp, http.StatusOK)
		var body authBody
		mustDecode(t, resp, &body)
		if body.Token == "" {
			t.Fatal("token must be non-empty")
		}
		if body.User.Email != "alice@example.com" {
			t.Fatalf("user.email: got %q", body.User.Email)
		}
	})

	t.Run("wrong password → 401", func(t *testing.T) {
		resp := e.loginRaw(t, "alice@example.com", "wrong-password")
		assertStatus(t, resp, http.StatusUnauthorized)
		resp.Body.Close()
	})

	t.Run("unknown email → 401", func(t *testing.T) {
		resp := e.loginRaw(t, "nobody@example.com", "password123")
		assertStatus(t, resp, http.StatusUnauthorized)
		resp.Body.Close()
	})
}

func TestAuth_ProtectedEndpoints(t *testing.T) {
	e := newEnv(t)

	endpoints := []struct{ method, path string }{
		{http.MethodGet, "/api/auth/me"},
		{http.MethodPost, "/api/auth/logout"},
		{http.MethodGet, "/api/projects"},
		{http.MethodPost, "/api/projects"},
		{http.MethodGet, "/api/users"},
	}
	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path+" no token → 401", func(t *testing.T) {
			resp := e.do(t, ep.method, ep.path, "", nil)
			assertStatus(t, resp, http.StatusUnauthorized)
			resp.Body.Close()
		})
		t.Run(ep.method+" "+ep.path+" bad token → 401", func(t *testing.T) {
			resp := e.do(t, ep.method, ep.path, "not-a-valid-jwt", nil)
			assertStatus(t, resp, http.StatusUnauthorized)
			resp.Body.Close()
		})
	}
}

func TestAuth_Logout(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")

	t.Run("token works before logout", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/auth/me", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		resp.Body.Close()
	})

	t.Run("logout returns 204", func(t *testing.T) {
		resp := e.do(t, http.MethodPost, "/api/auth/logout", alice.Token, nil)
		assertStatus(t, resp, http.StatusNoContent)
		resp.Body.Close()
	})

	t.Run("same token rejected after logout", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/auth/me", alice.Token, nil)
		assertStatus(t, resp, http.StatusUnauthorized)
		resp.Body.Close()
	})

	t.Run("second session remains valid after first session logout", func(t *testing.T) {

		resp := e.loginRaw(t, "alice@example.com", "password123")
		assertStatus(t, resp, http.StatusOK)
		var second authBody
		mustDecode(t, resp, &second)

		resp = e.do(t, http.MethodGet, "/api/auth/me", second.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		resp.Body.Close()
	})
}

func TestProjects_CRUD(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")

	t.Run("create returns 201 with id and name", func(t *testing.T) {
		p := e.createProject(t, alice.Token, "Alpha")
		if p["id"] == nil || p["id"] == "" {
			t.Fatal("expected non-empty id")
		}
		if p["name"] != "Alpha" {
			t.Fatalf("name: got %v", p["name"])
		}
		if p["owner_id"] == nil {
			t.Fatal("expected owner_id")
		}
	})

	t.Run("list includes owned project", func(t *testing.T) {
		e.createProject(t, alice.Token, "Beta")
		resp := e.do(t, http.MethodGet, "/api/projects", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		projects := body["projects"].([]any)
		if len(projects) < 1 {
			t.Fatalf("want ≥1 project, got %d", len(projects))
		}
	})

	t.Run("get project includes tasks array", func(t *testing.T) {
		p := e.createProject(t, alice.Token, "Gamma")
		pid := p["id"].(string)
		e.createTask(t, alice.Token, pid, "Task A", "todo")

		resp := e.do(t, http.MethodGet, "/api/projects/"+pid, alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var detail map[string]any
		mustDecode(t, resp, &detail)
		tasks, ok := detail["tasks"].([]any)
		if !ok || len(tasks) != 1 {
			t.Fatalf("want 1 task, got %v", detail["tasks"])
		}
	})

	t.Run("patch updates name", func(t *testing.T) {
		p := e.createProject(t, alice.Token, "Old Name")
		pid := p["id"].(string)

		resp := e.do(t, http.MethodPatch, "/api/projects/"+pid, alice.Token, map[string]string{"name": "New Name"})
		assertStatus(t, resp, http.StatusOK)
		var updated map[string]any
		mustDecode(t, resp, &updated)
		if updated["name"] != "New Name" {
			t.Fatalf("name: got %v", updated["name"])
		}
	})

	t.Run("delete returns 204 and project is gone", func(t *testing.T) {
		p := e.createProject(t, alice.Token, "To Delete")
		pid := p["id"].(string)

		resp := e.do(t, http.MethodDelete, "/api/projects/"+pid, alice.Token, nil)
		assertStatus(t, resp, http.StatusNoContent)
		resp.Body.Close()

		resp = e.do(t, http.MethodGet, "/api/projects/"+pid, alice.Token, nil)
		assertStatus(t, resp, http.StatusNotFound)
		resp.Body.Close()
	})
}

func TestProjects_Authorization(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")
	bob := e.register(t, "Bob", "bob@example.com", "password456")

	p := e.createProject(t, alice.Token, "Alice's Project")
	pid := p["id"].(string)

	t.Run("unrelated user gets 404 on GET project", func(t *testing.T) {

		resp := e.do(t, http.MethodGet, "/api/projects/"+pid, bob.Token, nil)
		assertStatus(t, resp, http.StatusNotFound)
		resp.Body.Close()
	})

	t.Run("non-owner gets 403 on PATCH", func(t *testing.T) {
		resp := e.do(t, http.MethodPatch, "/api/projects/"+pid, bob.Token, map[string]string{"name": "Hacked"})

		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("want 403 or 404, got %d — body: %s", resp.StatusCode, b)
		}
		resp.Body.Close()
	})

	t.Run("non-owner gets 403 on DELETE", func(t *testing.T) {
		resp := e.do(t, http.MethodDelete, "/api/projects/"+pid, bob.Token, nil)
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("want 403 or 404, got %d — body: %s", resp.StatusCode, b)
		}
		resp.Body.Close()
	})
}

func TestTasks_CreateAndList(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")
	bob := e.register(t, "Bob", "bob@example.com", "password456")

	p := e.createProject(t, alice.Token, "Filtered Project")
	pid := p["id"].(string)

	e.createTask(t, alice.Token, pid, "Todo 1", "todo")
	e.createTask(t, alice.Token, pid, "Todo 2", "todo")
	e.createTask(t, alice.Token, pid, "Done 1", "done")

	t.Run("list all tasks returns correct total", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		if body["total"].(float64) != 3 {
			t.Fatalf("total: want 3 got %v", body["total"])
		}
	})

	t.Run("?status=todo returns only todo tasks", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks?status=todo", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		tasks := body["tasks"].([]any)
		if len(tasks) != 2 {
			t.Fatalf("?status=todo: want 2 got %d", len(tasks))
		}
		for _, raw := range tasks {
			task := raw.(map[string]any)
			if task["status"] != "todo" {
				t.Fatalf("unexpected status %q in filtered results", task["status"])
			}
		}
	})

	t.Run("?status=done returns only done tasks", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks?status=done", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		tasks := body["tasks"].([]any)
		if len(tasks) != 1 {
			t.Fatalf("?status=done: want 1 got %d", len(tasks))
		}
	})

	t.Run("create with due_date in the past returns 400", func(t *testing.T) {
		past := time.Now().UTC().AddDate(0, 0, -2).Format("2006-01-02")
		resp := e.do(t, http.MethodPost, "/api/projects/"+pid+"/tasks", alice.Token, map[string]string{
			"title": "Stale due", "due_date": past,
		})
		assertStatus(t, resp, http.StatusBadRequest)
		resp.Body.Close()
	})

	t.Run("create with due_date today returns 201", func(t *testing.T) {
		today := time.Now().UTC().Format("2006-01-02")
		resp := e.do(t, http.MethodPost, "/api/projects/"+pid+"/tasks", alice.Token, map[string]string{
			"title": "Due today", "due_date": today,
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()
	})

	t.Run("?assignee= filters by assignee_id", func(t *testing.T) {

		bobID := bob.User.ID
		resp := e.do(t, http.MethodPost, "/api/projects/"+pid+"/tasks", alice.Token, map[string]string{
			"title": "Bob's Task", "assignee_id": bobID,
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()

		resp = e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks?assignee="+bobID, alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		tasks := body["tasks"].([]any)
		if len(tasks) != 1 {
			t.Fatalf("?assignee=bob: want 1 got %d", len(tasks))
		}
	})

	t.Run("invalid status filter returns 400", func(t *testing.T) {
		resp := e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks?status=invalid", alice.Token, nil)
		assertStatus(t, resp, http.StatusBadRequest)
		resp.Body.Close()
	})
}

func TestTasks_PatchAndDelete(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")
	bob := e.register(t, "Bob", "bob@example.com", "password456")

	p := e.createProject(t, alice.Token, "Task Ops Project")
	pid := p["id"].(string)
	task := e.createTask(t, alice.Token, pid, "Original Title", "todo")
	tid := task["id"].(string)

	t.Run("patch title and status", func(t *testing.T) {
		resp := e.do(t, http.MethodPatch, "/api/tasks/"+tid, alice.Token, map[string]string{
			"title": "Updated Title", "status": "in_progress",
		})
		assertStatus(t, resp, http.StatusOK)
		var updated map[string]any
		mustDecode(t, resp, &updated)
		if updated["title"] != "Updated Title" {
			t.Fatalf("title: got %v", updated["title"])
		}
		if updated["status"] != "in_progress" {
			t.Fatalf("status: got %v", updated["status"])
		}
	})

	t.Run("patch with invalid status returns 400", func(t *testing.T) {
		resp := e.do(t, http.MethodPatch, "/api/tasks/"+tid, alice.Token, map[string]string{
			"status": "not-a-valid-status",
		})
		assertStatus(t, resp, http.StatusBadRequest)
		resp.Body.Close()
	})

	t.Run("project collaborator may patch task they did not create and are not assignee on", func(t *testing.T) {
		resp := e.do(t, http.MethodPost, "/api/projects/"+pid+"/tasks", alice.Token, map[string]string{
			"title": "Bob lane", "assignee_id": bob.User.ID,
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()

		resp = e.do(t, http.MethodPatch, "/api/tasks/"+tid, bob.Token, map[string]string{
			"title": "Bob updated Alice task",
		})
		assertStatus(t, resp, http.StatusOK)
		var updated map[string]any
		mustDecode(t, resp, &updated)
		if updated["title"] != "Bob updated Alice task" {
			t.Fatalf("title: got %v", updated["title"])
		}
	})

	t.Run("non-owner non-creator delete → 403 or 404", func(t *testing.T) {
		resp := e.do(t, http.MethodDelete, "/api/tasks/"+tid, bob.Token, nil)
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("want 403 or 404, got %d — body: %s", resp.StatusCode, b)
		}
		resp.Body.Close()
	})

	t.Run("owner deletes task → 204, task is gone", func(t *testing.T) {
		resp := e.do(t, http.MethodDelete, "/api/tasks/"+tid, alice.Token, nil)
		assertStatus(t, resp, http.StatusNoContent)
		resp.Body.Close()

		resp = e.do(t, http.MethodGet, "/api/projects/"+pid+"/tasks", alice.Token, nil)
		assertStatus(t, resp, http.StatusOK)
		var body map[string]any
		mustDecode(t, resp, &body)
		if body["total"].(float64) != 0 {
			t.Fatalf("total after delete: want 0 got %v", body["total"])
		}
	})
}

func TestProjects_Stats(t *testing.T) {
	e := newEnv(t)
	alice := e.register(t, "Alice", "alice@example.com", "password123")
	p := e.createProject(t, alice.Token, "Stats Project")
	pid := p["id"].(string)

	e.createTask(t, alice.Token, pid, "T1", "todo")
	e.createTask(t, alice.Token, pid, "T2", "todo")
	e.createTask(t, alice.Token, pid, "T3", "done")

	resp := e.do(t, http.MethodGet, "/api/projects/"+pid+"/stats", alice.Token, nil)
	assertStatus(t, resp, http.StatusOK)
	var body map[string]any
	mustDecode(t, resp, &body)

	byStatus, ok := body["by_status"].(map[string]any)
	if !ok {
		t.Fatalf("by_status missing or wrong type: %v", body)
	}
	if byStatus["todo"].(float64) != 2 {
		t.Fatalf("by_status.todo: want 2 got %v", byStatus["todo"])
	}
	if byStatus["done"].(float64) != 1 {
		t.Fatalf("by_status.done: want 1 got %v", byStatus["done"])
	}
}
