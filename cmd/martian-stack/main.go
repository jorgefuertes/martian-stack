package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// CLI flags for non-interactive mode.
var (
	flagName   = flag.String("name", "", "Project name")
	flagModule = flag.String("module", "", "Go module path")
	flagOutput = flag.String("output", "", "Output directory")
	flagDB     = flag.String("db", "sqlite", "Database: sqlite, postgres, mysql, none")
	flagCache  = flag.String("cache", "memory", "Cache: memory, redis")
	flagAuth   = flag.Bool("auth", true, "Include JWT authentication")
	flagAdmin  = flag.Bool("admin", true, "Include admin panel")
	flagMw     = flag.String("middlewares", "cors,logging,security,recovery", "Comma-separated middlewares")
	flagYes    = flag.Bool("y", false, "Skip confirmation (non-interactive)")
)

// ProjectConfig holds all user choices for project generation.
type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	OutputDir   string

	Database string // sqlite, postgres, mysql, none
	Cache    string // memory, redis

	MwCORS            bool
	MwLogging         bool
	MwSecurityHeaders bool
	MwRateLimit       bool
	MwRecovery        bool
	MwTimeout         bool

	HasAuth  bool
	HasAdmin bool

	// Computed
	HasDatabase bool
	HasRedis    bool

	DefaultHost    string
	DefaultPort    string
	DefaultTimeout int
}

func main() {
	flag.Parse()

	printBanner()

	var cfg ProjectConfig
	var err error

	if *flagYes {
		cfg, err = configFromFlags()
	} else {
		cfg, err = runForm()
	}

	if err != nil {
		if err.Error() == "user aborted" {
			fmt.Println("\nAborted.")
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Computed fields
	cfg.HasDatabase = cfg.Database != "none"
	cfg.HasRedis = cfg.Cache == "redis"
	cfg.DefaultHost = "localhost"
	cfg.DefaultPort = "8080"
	cfg.DefaultTimeout = 15

	// Auth requires a database
	if !cfg.HasDatabase {
		cfg.HasAuth = false
		cfg.HasAdmin = false
	}

	if !*flagYes {
		if !printSummary(cfg) {
			fmt.Println("\nAborted.")
			os.Exit(0)
		}
	}

	if err := generate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "\nError generating project: %v\n", err)
		os.Exit(1)
	}

	printSuccess(cfg)
}

func configFromFlags() (ProjectConfig, error) {
	if *flagName == "" || *flagModule == "" || *flagOutput == "" {
		return ProjectConfig{}, fmt.Errorf("--name, --module, and --output are required in non-interactive mode")
	}

	cfg := ProjectConfig{
		ProjectName: *flagName,
		ModulePath:  *flagModule,
		OutputDir:   *flagOutput,
		Database:    *flagDB,
		Cache:       *flagCache,
		HasAuth:     *flagAuth,
		HasAdmin:    *flagAdmin,
	}

	for _, mw := range strings.Split(*flagMw, ",") {
		switch strings.TrimSpace(mw) {
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

	return cfg, nil
}

// --- Banner ---

func printBanner() {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6600")).
		Render("MARTIAN STACK")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("Project Generator")

	fmt.Printf("\n  %s  %s\n\n", title, subtitle)
}

// --- Form ---

func runForm() (ProjectConfig, error) {
	var cfg ProjectConfig

	// Step 1: Project basics
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Short name for your project").
				Placeholder("my-app").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("project name is required")
					}
					return nil
				}).
				Value(&cfg.ProjectName),
			huh.NewInput().
				Title("Go module path").
				Description("e.g. github.com/user/project").
				Placeholder("github.com/user/my-app").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("module path is required")
					}
					return nil
				}).
				Value(&cfg.ModulePath),
			huh.NewInput().
				Title("Output directory").
				Description("Where to create the project").
				Placeholder("./my-app").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("output directory is required")
					}
					return nil
				}).
				Value(&cfg.OutputDir),
		).Title("Project Basics"),
	).Run()
	if err != nil {
		return cfg, err
	}

	// Step 2: Infrastructure
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database").
				Options(
					huh.NewOption("SQLite", "sqlite"),
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("MySQL", "mysql"),
					huh.NewOption("None", "none"),
				).
				Value(&cfg.Database),
			huh.NewSelect[string]().
				Title("Cache").
				Options(
					huh.NewOption("Memory", "memory"),
					huh.NewOption("Redis", "redis"),
				).
				Value(&cfg.Cache),
		).Title("Infrastructure"),
	).Run()
	if err != nil {
		return cfg, err
	}

	// Step 3: Middlewares
	var selectedMw []string
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Middlewares").
				Options(
					huh.NewOption("CORS", "cors").Selected(true),
					huh.NewOption("Logging", "logging").Selected(true),
					huh.NewOption("Security Headers", "security").Selected(true),
					huh.NewOption("Rate Limiting", "ratelimit"),
					huh.NewOption("Recovery", "recovery").Selected(true),
					huh.NewOption("Timeout", "timeout"),
				).
				Value(&selectedMw),
		).Title("Middlewares"),
	).Run()
	if err != nil {
		return cfg, err
	}

	for _, mw := range selectedMw {
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

	// Step 4: Authentication (only if database selected)
	if cfg.Database != "none" {
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Include JWT authentication?").
					Value(&cfg.HasAuth),
			).Title("Authentication"),
		).Run()
		if err != nil {
			return cfg, err
		}

		if cfg.HasAuth {
			err = huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Include admin panel with user CRUD?").
						Value(&cfg.HasAdmin),
				).Title("Admin Panel"),
			).Run()
			if err != nil {
				return cfg, err
			}
		}
	}

	return cfg, nil
}

// --- Summary ---

func printSummary(cfg ProjectConfig) bool {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6600"))

	labelStyle := lipgloss.NewStyle().
		Width(20).
		Foreground(lipgloss.Color("#888888"))

	valueStyle := lipgloss.NewStyle().
		Bold(true)

	fmt.Println()
	fmt.Println(headerStyle.Render("  Project Summary"))
	fmt.Println(headerStyle.Render("  ───────────────"))

	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	printRow("Name", cfg.ProjectName)
	printRow("Module", cfg.ModulePath)
	printRow("Directory", cfg.OutputDir)
	printRow("Database", cfg.Database)
	printRow("Cache", cfg.Cache)

	var mws []string
	if cfg.MwRecovery {
		mws = append(mws, "Recovery")
	}
	if cfg.MwSecurityHeaders {
		mws = append(mws, "Security Headers")
	}
	if cfg.MwRateLimit {
		mws = append(mws, "Rate Limiting")
	}
	if cfg.MwCORS {
		mws = append(mws, "CORS")
	}
	if cfg.MwTimeout {
		mws = append(mws, "Timeout")
	}
	if cfg.MwLogging {
		mws = append(mws, "Logging")
	}
	if len(mws) > 0 {
		printRow("Middlewares", strings.Join(mws, ", "))
	}

	if cfg.HasAuth {
		printRow("JWT Auth", "yes")
	}
	if cfg.HasAdmin {
		printRow("Admin Panel", "yes")
	}
	fmt.Println()

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate project?").
				Value(&confirm),
		),
	).Run()
	if err != nil {
		return false
	}

	return confirm
}

// --- Generation ---

type fileEntry struct {
	path     string
	template string
	cond     bool
}

func generate(cfg ProjectConfig) error {
	files := []fileEntry{
		{"main.go", tplMain, true},
		{"go.mod", tplGoMod, true},
		{"Makefile", tplMakefile, true},
		{".env.example", tplEnvExample, true},
		{".gitignore", tplGitignore, true},
		{"handlers/routes.go", tplRoutes, true},
		{"database/database.go", tplDatabase, cfg.HasDatabase},
		{"database/migrations/migrations.go", tplMigrations, cfg.HasDatabase},
		{"database/migrations/001_initial.go", tplMigration001Accounts, cfg.HasDatabase && cfg.HasAuth},
		{"database/migrations/001_initial.go", tplMigration001Example, cfg.HasDatabase && !cfg.HasAuth},
		{"database/migrations/002_token_tables.go", tplMigration002, cfg.HasDatabase && cfg.HasAuth},
		{"handlers/auth.go", tplAuthRoutes, cfg.HasAuth},
		{"handlers/admin.go", tplAdminRoutes, cfg.HasAdmin},
	}

	for _, f := range files {
		if !f.cond {
			continue
		}
		outPath := filepath.Join(cfg.OutputDir, f.path)
		if err := renderAndWrite(outPath, f.template, cfg); err != nil {
			return fmt.Errorf("generating %s: %w", f.path, err)
		}
		fmt.Printf("  created %s\n", f.path)
	}

	return nil
}

func renderAndWrite(path, tplStr string, data ProjectConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpl, err := template.New("").Parse(tplStr)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// --- Success ---

func printSuccess(cfg ProjectConfig) {
	ok := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00CC00"))

	fmt.Println()
	fmt.Println(ok.Render("  Project generated successfully!"))
	fmt.Println()

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	fmt.Println(hint.Render("  Next steps:"))
	fmt.Printf("    cd %s\n", cfg.OutputDir)
	fmt.Println("    go mod tidy")
	fmt.Println("    go run .")
	fmt.Println()
}
