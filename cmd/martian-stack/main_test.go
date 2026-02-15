package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseConfig returns a ProjectConfig with computed fields set.
func baseConfig(db, cache string, auth, admin bool, mws ...string) ProjectConfig {
	cfg := ProjectConfig{
		ProjectName:    "testapp",
		ModulePath:     "github.com/test/testapp",
		Database:       db,
		Cache:          cache,
		HasAuth:        auth,
		HasAdmin:       admin,
		HasDatabase:    db != "none",
		HasRedis:       cache == "redis",
		DefaultHost:    "localhost",
		DefaultPort:    "8080",
		DefaultTimeout: 15,
	}

	for _, mw := range mws {
		switch mw {
		case "cors":
			cfg.MwCORS = true
		case "logging":
			cfg.MwLogging = true
		case "security":
			cfg.MwSecurityHeaders = true
		case "ratelimit":
			cfg.MwRateLimit = true
		case "recovery":
			cfg.MwRecovery = true
		case "timeout":
			cfg.MwTimeout = true
		}
	}

	if !cfg.HasDatabase {
		cfg.HasAuth = false
		cfg.HasAdmin = false
	}

	return cfg
}

func renderTemplate(t *testing.T, tplStr string, cfg ProjectConfig) string {
	t.Helper()
	tmpl, err := template.New("").Parse(tplStr)
	require.NoError(t, err, "template should parse")
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, cfg), "template should render")
	return buf.String()
}

// allTemplates returns every template constant paired with a name.
func allTemplates() []struct {
	name string
	tpl  string
} {
	return []struct {
		name string
		tpl  string
	}{
		{"tplMain", tplMain},
		{"tplGoMod", tplGoMod},
		{"tplMakefile", tplMakefile},
		{"tplEnvExample", tplEnvExample},
		{"tplGitignore", tplGitignore},
		{"tplRoutes", tplRoutes},
		{"tplDatabase", tplDatabase},
		{"tplMigrations", tplMigrations},
		{"tplMigration001Accounts", tplMigration001Accounts},
		{"tplMigration001Example", tplMigration001Example},
		{"tplMigration002", tplMigration002},
		{"tplAuthRoutes", tplAuthRoutes},
		{"tplAdminRoutes", tplAdminRoutes},
	}
}

func TestTemplateParsing(t *testing.T) {
	for _, tt := range allTemplates() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := template.New(tt.name).Parse(tt.tpl)
			require.NoError(t, err)
		})
	}
}

func TestTemplateRendering(t *testing.T) {
	configs := []struct {
		name string
		cfg  ProjectConfig
	}{
		{
			"full-sqlite",
			baseConfig(
				"sqlite",
				"memory",
				true,
				true,
				"cors",
				"logging",
				"security",
				"recovery",
				"ratelimit",
				"timeout",
			),
		},
		{"full-postgres-redis", baseConfig("postgres", "redis", true, true, "cors", "logging", "security", "recovery")},
		{"full-mysql", baseConfig("mysql", "memory", true, false, "cors", "logging")},
		{"db-no-auth", baseConfig("sqlite", "memory", false, false, "recovery")},
		{"no-db", baseConfig("none", "memory", false, false, "cors", "logging", "recovery")},
		{"no-db-redis", baseConfig("none", "redis", false, false)},
		{"minimal", baseConfig("none", "memory", false, false)},
	}

	alwaysTemplates := []struct {
		name string
		tpl  string
	}{
		{"tplMain", tplMain},
		{"tplGoMod", tplGoMod},
		{"tplMakefile", tplMakefile},
		{"tplEnvExample", tplEnvExample},
		{"tplGitignore", tplGitignore},
		{"tplRoutes", tplRoutes},
	}

	dbTemplates := []struct {
		name string
		tpl  string
	}{
		{"tplDatabase", tplDatabase},
		{"tplMigrations", tplMigrations},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			for _, tt := range alwaysTemplates {
				renderTemplate(t, tt.tpl, tc.cfg)
			}

			if tc.cfg.HasDatabase {
				for _, tt := range dbTemplates {
					renderTemplate(t, tt.tpl, tc.cfg)
				}
				if tc.cfg.HasAuth {
					renderTemplate(t, tplMigration001Accounts, tc.cfg)
					renderTemplate(t, tplMigration002, tc.cfg)
					renderTemplate(t, tplAuthRoutes, tc.cfg)
				} else {
					renderTemplate(t, tplMigration001Example, tc.cfg)
				}
				if tc.cfg.HasAdmin {
					renderTemplate(t, tplAdminRoutes, tc.cfg)
				}
			}
		})
	}
}

func TestTemplateContentMain(t *testing.T) {
	t.Run("imports database when db=sqlite", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("sqlite", "memory", true, false, "cors"))

		assert.Contains(t, out, `"github.com/test/testapp/database"`)
		assert.Contains(t, out, `"github.com/test/testapp/database/migrations"`)
		assert.Contains(t, out, `database.Connect()`)
		assert.Contains(t, out, `migrations.Run(db)`)
	})

	t.Run("omits database when db=none", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("none", "memory", false, false))

		assert.NotContains(t, out, `database.Connect()`)
		assert.NotContains(t, out, `migrations.Run`)
	})

	t.Run("uses redis cache when cache=redis", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("none", "redis", false, false))

		assert.Contains(t, out, `cache/redis`)
		assert.NotContains(t, out, `cache/memory`)
	})

	t.Run("uses memory cache when cache=memory", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("none", "memory", false, false))

		assert.Contains(t, out, `cache/memory`)
		assert.NotContains(t, out, `cache/redis`)
	})

	t.Run("includes JWT when auth enabled", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("sqlite", "memory", true, false))

		assert.Contains(t, out, `jwt.DefaultConfig`)
		assert.Contains(t, out, `auth.NewHandlers`)
		assert.Contains(t, out, `auth.NewMiddleware`)
		assert.Contains(t, out, `handlers.RegisterAuthRoutes`)
	})

	t.Run("includes admin routes when admin enabled", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("sqlite", "memory", true, true))

		assert.Contains(t, out, `handlers.RegisterAdminRoutes`)
	})

	t.Run("includes timeout middleware and time import", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("none", "memory", false, false, "timeout"))

		assert.Contains(t, out, `"time"`)
		assert.Contains(t, out, `middleware.NewTimeout`)
	})

	t.Run("omits time import without timeout middleware", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig("none", "memory", false, false, "cors"))

		assert.NotContains(t, out, `"time"`)
	})

	t.Run("includes all selected middlewares", func(t *testing.T) {
		out := renderTemplate(t, tplMain, baseConfig(
			"none", "memory", false, false,
			"cors", "logging", "security", "ratelimit", "recovery", "timeout",
		))

		assert.Contains(t, out, `middleware.NewRecovery()`)
		assert.Contains(t, out, `middleware.NewSecurityHeaders()`)
		assert.Contains(t, out, `middleware.NewRateLimit`)
		assert.Contains(t, out, `middleware.NewCors`)
		assert.Contains(t, out, `middleware.NewTimeout`)
		assert.Contains(t, out, `middleware.NewLog(l)`)
	})
}

func TestTemplateContentDatabase(t *testing.T) {
	drivers := map[string]string{
		"sqlite":   `_ "modernc.org/sqlite"`,
		"postgres": `_ "github.com/jackc/pgx/v5/stdlib"`,
		"mysql":    `_ "github.com/go-sql-driver/mysql"`,
	}
	for db, expectedImport := range drivers {
		t.Run(db, func(t *testing.T) {
			out := renderTemplate(t, tplDatabase, baseConfig(db, "memory", false, false))
			assert.Contains(t, out, expectedImport)
		})
	}
}

func TestTemplateContentEnvExample(t *testing.T) {
	t.Run("includes redis section when redis", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("none", "redis", false, false))
		assert.Contains(t, out, "REDIS_HOST")
	})

	t.Run("excludes redis section when memory", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("none", "memory", false, false))
		assert.NotContains(t, out, "REDIS_HOST")
	})

	t.Run("includes JWT section when auth", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("sqlite", "memory", true, false))
		assert.Contains(t, out, "JWT_SECRET")
	})

	t.Run("excludes JWT section when no auth", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("none", "memory", false, false))
		assert.NotContains(t, out, "JWT_SECRET")
	})

	t.Run("includes sqlite DSN", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("sqlite", "memory", false, false))
		assert.Contains(t, out, "DB_DRIVER=sqlite3")
	})

	t.Run("includes postgres DSN", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("postgres", "memory", false, false))
		assert.Contains(t, out, "DB_DRIVER=postgres")
	})

	t.Run("includes mysql DSN", func(t *testing.T) {
		out := renderTemplate(t, tplEnvExample, baseConfig("mysql", "memory", false, false))
		assert.Contains(t, out, "DB_DRIVER=mysql")
	})
}

func TestGenerateCreatesFiles(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ProjectConfig
		expected []string
		absent   []string
	}{
		{
			"full project",
			baseConfig("sqlite", "memory", true, true, "cors", "recovery"),
			[]string{
				"main.go", "go.mod", "Makefile", ".env.example", ".gitignore",
				"handlers/routes.go", "handlers/auth.go", "handlers/admin.go",
				"database/database.go", "database/migrations/migrations.go",
				"database/migrations/001_initial.go", "database/migrations/002_token_tables.go",
			},
			nil,
		},
		{
			"no database",
			baseConfig("none", "memory", false, false, "cors"),
			[]string{"main.go", "go.mod", "Makefile", ".env.example", ".gitignore", "handlers/routes.go"},
			[]string{"database/database.go", "handlers/auth.go", "handlers/admin.go"},
		},
		{
			"db without auth",
			baseConfig("postgres", "redis", false, false),
			[]string{
				"main.go", "go.mod", "handlers/routes.go",
				"database/database.go", "database/migrations/migrations.go",
				"database/migrations/001_initial.go",
			},
			[]string{"handlers/auth.go", "handlers/admin.go", "database/migrations/002_token_tables.go"},
		},
		{
			"auth without admin",
			baseConfig("mysql", "memory", true, false, "cors"),
			[]string{"handlers/auth.go", "database/migrations/002_token_tables.go"},
			[]string{"handlers/admin.go"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.cfg.OutputDir = dir

			require.NoError(t, generate(tc.cfg))

			for _, f := range tc.expected {
				path := filepath.Join(dir, f)
				info, err := os.Stat(path)
				require.NoError(t, err, "expected file %s to exist", f)
				assert.Greater(t, info.Size(), int64(0), "file %s should not be empty", f)
			}

			for _, f := range tc.absent {
				path := filepath.Join(dir, f)
				_, err := os.Stat(path)
				assert.True(t, os.IsNotExist(err), "file %s should not exist", f)
			}
		})
	}
}

func TestGeneratedProjectsCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation tests in short mode")
	}

	frameworkRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(frameworkRoot, "go.mod"))

	tests := []struct {
		name string
		cfg  ProjectConfig
	}{
		{
			"sqlite-full",
			baseConfig(
				"sqlite",
				"memory",
				true,
				true,
				"cors",
				"logging",
				"security",
				"recovery",
				"ratelimit",
				"timeout",
			),
		},
		{"postgres-redis-auth", baseConfig("postgres", "redis", true, false, "cors", "logging", "recovery")},
		{"mysql-auth-admin", baseConfig("mysql", "memory", true, true, "cors", "recovery")},
		{"sqlite-no-auth", baseConfig("sqlite", "memory", false, false, "cors", "logging")},
		{"minimal", baseConfig("none", "memory", false, false, "cors", "logging", "recovery")},
		{"none-redis", baseConfig("none", "redis", false, false, "recovery")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.cfg.OutputDir = dir

			require.NoError(t, generate(tc.cfg))

			// Add replace directive for local framework
			cmd := exec.Command(
				"go", "mod", "edit",
				"-replace=github.com/jorgefuertes/martian-stack="+frameworkRoot,
			)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "go mod edit failed: %s", out)

			// go mod tidy
			cmd = exec.Command("go", "mod", "tidy")
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
			require.NoError(t, err, "go mod tidy failed: %s", out)

			// go build
			cmd = exec.Command("go", "build", ".")
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
			require.NoError(t, err, "go build failed: %s", out)
		})
	}
}
