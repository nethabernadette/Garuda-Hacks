package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"garuda-hacks/backend/auth"
	"garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/internal/ai"
	"garuda-hacks/backend/internal/chat"
	"garuda-hacks/backend/internal/document"
	"garuda-hacks/backend/notifications"
	"garuda-hacks/backend/offer"
	"garuda-hacks/backend/organizations"
	"garuda-hacks/backend/posts"
	"garuda-hacks/backend/product"
	"garuda-hacks/backend/users"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/go-sqlite"
	gsqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// SQLiteDriverWrapper wraps the glebarez driver to intercept connection creation.
type SQLiteDriverWrapper struct {
	parent *sqlite.Driver
}

func (d *SQLiteDriverWrapper) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}
	return &connWrapper{conn: conn}, nil
}

// connWrapper wraps driver.Conn to intercept SQL execution and prepare statements.
type connWrapper struct {
	conn driver.Conn
}

func (c *connWrapper) Prepare(query string) (driver.Stmt, error) {
	query = cleanSQL(query)
	return c.conn.Prepare(query)
}

func (c *connWrapper) Close() error {
	return c.conn.Close()
}

func (c *connWrapper) Begin() (driver.Tx, error) {
	return c.conn.Begin()
}

func (c *connWrapper) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	query = cleanSQL(query)
	if pc, ok := c.conn.(driver.ConnPrepareContext); ok {
		return pc.PrepareContext(ctx, query)
	}
	return c.conn.Prepare(query)
}

func (c *connWrapper) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if bt, ok := c.conn.(driver.ConnBeginTx); ok {
		return bt.BeginTx(ctx, opts)
	}
	return c.conn.Begin()
}

func (c *connWrapper) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	query = cleanSQL(query)
	if ec, ok := c.conn.(driver.ExecerContext); ok {
		return ec.ExecContext(ctx, query, args)
	}
	if execer, ok := c.conn.(driver.Execer); ok {
		dargs := make([]driver.Value, len(args))
		for i, v := range args {
			dargs[i] = v.Value
		}
		return execer.Exec(query, dargs)
	}
	return nil, driver.ErrSkip
}

func (c *connWrapper) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	query = cleanSQL(query)
	if qc, ok := c.conn.(driver.QueryerContext); ok {
		return qc.QueryContext(ctx, query, args)
	}
	if queryer, ok := c.conn.(driver.Queryer); ok {
		dargs := make([]driver.Value, len(args))
		for i, v := range args {
			dargs[i] = v.Value
		}
		return queryer.Query(query, dargs)
	}
	return nil, driver.ErrSkip
}

func (c *connWrapper) Ping(ctx context.Context) error {
	if pinger, ok := c.conn.(driver.Pinger); ok {
		return pinger.Ping(ctx)
	}
	return nil
}

// cleanSQL replaces PostgreSQL-specific defaults with SQLite equivalents.
func cleanSQL(query string) string {
	// Replace "DEFAULT gen_random_uuid()" with SQLite-compatible RFC4122 UUID generator.
	uuidExpr := "DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))))"
	query = strings.ReplaceAll(query, "DEFAULT gen_random_uuid()", uuidExpr)
	query = strings.ReplaceAll(query, "default gen_random_uuid()", uuidExpr)
	return query
}

// Register the custom now() function and our driver wrapper.
func init() {
	err := sqlite.RegisterDeterministicScalarFunction("now", 0, func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		return time.Now().UTC().Format("2006-01-02 15:04:05.999999999-07:00"), nil
	})
	if err != nil {
		log.Fatalf("Failed to register now() function: %v", err)
	}

	sql.Register("sqlite_custom_driver", &SQLiteDriverWrapper{
		parent: &sqlite.Driver{},
	})
}

// postsNotificationAdapter maps notifications.Service to posts.NotificationCreator.
type postsNotificationAdapter struct {
	notifService notifications.Service
}

func (a *postsNotificationAdapter) CreateUnique(ctx context.Context, userID string, notifType string, title string, message string, refType string, refID string) error {
	return a.notifService.CreateUnique(ctx, userID, notifType, title, message, refType, refID)
}

// CombinedHandler routes requests to standard library Mux or Gin Engine.
type CombinedHandler struct {
	Mux *http.ServeMux
	Gin *gin.Engine
}

func (h *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/posts") ||
		strings.HasPrefix(path, "/notifications") ||
		strings.HasPrefix(path, "/products") ||
		strings.HasPrefix(path, "/producer") ||
		strings.HasPrefix(path, "/offers") ||
		strings.HasPrefix(path, "/demand-groups") {
		h.Gin.ServeHTTP(w, r)
		return
	}
	h.Mux.ServeHTTP(w, r)
}

// corsMiddleware wraps the combined handler to add CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ginLaxAuth parses the token if present and sets user details in the Gin context.
func ginLaxAuth(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := tokenManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Inject into Gin Context
		c.Set("CurrentUserID", claims.UserID)
		c.Set("CurrentUserRole", string(claims.Role))

		c.Next()
	}
}

func main() {
	log.Println("Starting FoodLink AI local backend...")

	// Load env from project root or local folder
	loadEnv("../.env", ".env")

	// Default env values for fallback
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "super-secret-key-for-local-hackathon")
	}
	if os.Getenv("JWT_ACCESS_TOKEN_TTL_SECONDS") == "" {
		os.Setenv("JWT_ACCESS_TOKEN_TTL_SECONDS", "3600")
	}
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8080")
	}

	// 1. Initialize SQLite Database using our custom SQL-intercepting driver
	db, err := gorm.Open(gsqlite.Dialector{
		DriverName: "sqlite_custom_driver",
		DSN:        "garuda.db",
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite database: %v", err)
	}
	log.Println("Database connection established successfully.")

	// Enable SQLite foreign keys
	db.Exec("PRAGMA foreign_keys = ON;")

	// 2. Perform Migrations
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(
		&users.User{},
		&users.UserProfile{},
		&users.NIBVerification{},
		&organizations.Organization{},
		&organizations.OrganizationMember{},
		&auth.RefreshToken{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate users/organizations schemas: %v", err)
	}

	if err := posts.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate posts schema: %v", err)
	}

	if err := notifications.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate notifications schema: %v", err)
	}

	if err := agreement.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate agreement schema: %v", err)
	}

	if err := chat.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate chat schema: %v", err)
	}

	if err := product.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate product schema: %v", err)
	}

	if err := offer.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate offer schema: %v", err)
	}

	if err := ai.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate AI schema: %v", err)
	}

	log.Println("Migrations executed successfully.")

	// 3. Initialize Services and Handlers

	// Auth Initialization
	authRepo := auth.NewGormRepository(db)
	authConfig, err := auth.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load auth config: %v", err)
	}
	tokenManager := auth.NewTokenManager(authConfig.JWTSecret, authConfig.AccessTokenTTL)
	authService := auth.NewService(authRepo, tokenManager)
	authHandler := auth.NewHandler(authService)

	// Authenticate Middleware for standard library handlers
	authenticateMdw := auth.Authenticate(tokenManager)

	// Users Initialization
	claimsExtractor := func(ctx context.Context) (users.Principal, bool) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			return users.Principal{}, false
		}
		return users.Principal{
			UserID: claims.UserID,
			Role:   claims.Role,
		}, true
	}
	usersRepo := users.NewGormRepository(db)
	usersService := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersService, claimsExtractor)

	// Organizations Initialization
	orgRepo := organizations.NewGormRepository(db)
	orgService := organizations.NewService(orgRepo)
	orgHandler := organizations.NewHandler(orgService)

	// 4. Set up standard library routing
	mux := http.NewServeMux()

	auth.RegisterRoutes(mux, authHandler)
	users.RegisterRoutes(mux, usersHandler, authenticateMdw)
	organizations.RegisterRoutes(mux, orgHandler, authenticateMdw, orgRepo)
	agreement.RegisterRoutes(mux, db, authenticateMdw)
	chat.RegisterRoutes(mux, db, authenticateMdw)
	document.RegisterRoutes(mux, db, authenticateMdw)
	ai.RegisterRoutes(mux, db, authenticateMdw)

	// 5. Set up Gin engine and routing
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	// Global CORS is already handled by our outer http middleware.
	// We add Lax Authentication to Gin routes to parse Bearer tokens.
	router.Use(ginLaxAuth(tokenManager))

	// Register Gin-based package routes
	notifService := notifications.RegisterRoutes(router, db)
	postsAdapter := &postsNotificationAdapter{notifService: notifService}
	posts.RegisterRoutes(router, db, postsAdapter)
	product.RegisterRoutes(router, db)
	offer.RegisterRoutes(router, db)

	// 6. Bind standard library and Gin together under combined handler
	combined := &CombinedHandler{
		Mux: mux,
		Gin: router,
	}

	// 7. Start Server
	port := os.Getenv("PORT")
	serverAddr := ":" + port
	log.Printf("Server listening on http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, corsMiddleware(combined)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// loadEnv loads environment variables from a .env file.
func loadEnv(filenames ...string) {
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer file.Close()

		var lines []string
		buf := make([]byte, 65536)
		n, err := file.Read(buf)
		if err != nil && n == 0 {
			continue
		}
		content := string(buf[:n])
		lines = strings.Split(content, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				os.Setenv(key, value)
			}
		}
	}
}
