package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/catatsuy/private-isu/webapp/golang/cache"
	"github.com/catatsuy/private-isu/webapp/golang/handlers"
	"github.com/catatsuy/private-isu/webapp/golang/middleware"
	"github.com/catatsuy/private-isu/webapp/golang/utils"
	"github.com/bradfitz/gomemcache/memcache"
	gsm "github.com/bradleypeabody/gorilla-sessions-memcache"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	db    *sqlx.DB
	store *gsm.MemcacheStore
)

func getTemplPath(filename string) string {
	return path.Join("templates", filename)
}

func dbInitialize() {
	sqls := []string{
		"DELETE FROM users WHERE id > 1000",
		"DELETE FROM posts WHERE id > 10000",
		"DELETE FROM comments WHERE id > 100000",
		"UPDATE users SET del_flg = 0",
		"UPDATE users SET del_flg = 1 WHERE id % 50 = 0",
	}

	for _, sql := range sqls {
		db.Exec(sql)
	}

	// Clear all caches on initialization
	if client := cache.GetClient(); client != nil {
		client.FlushAll()
	}
}

func main() {
	host := os.Getenv("ISUCONP_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("ISUCONP_DB_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable ISUCONP_DB_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("ISUCONP_DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("ISUCONP_DB_PASSWORD")
	dbname := os.Getenv("ISUCONP_DB_NAME")
	if dbname == "" {
		dbname = "isuconp"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	db, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	defer db.Close()

	memdAddr := os.Getenv("ISUCONP_MEMCACHED_ADDRESS")
	if memdAddr == "" {
		memdAddr = "localhost:11211"
	}

	// Initialize cache
	cache.Initialize(memdAddr)

	memcacheClient := cache.GetClient()
	store = gsm.NewMemcacheStore(memcacheClient, "iscogram_", []byte("sendagaya"))

	// Initialize handlers and utils
	handlers.InitHandlers(db, store)
	utils.InitSession(db, store)

	// Start metrics collector
	go middleware.MetricsCollector()

	r := chi.NewRouter()

	// Add metrics middleware
	r.Use(middleware.MetricsMiddleware)

	// Routes
	r.Get("/initialize", dbInitialize)
	r.Get("/login", handlers.GetLogin)
	r.Post("/login", handlers.PostLogin)
	r.Get("/register", handlers.GetRegister)
	r.Post("/register", handlers.PostRegister)
	r.Get("/logout", handlers.GetLogout)
	r.Get("/", handlers.GetIndex)
	r.Post("/", handlers.PostIndex)
	r.Get("/posts", handlers.GetPosts)
	r.Get("/posts/{id}", handlers.GetPostsID)
	r.Get("/image/{id}.{ext}", handlers.GetImage)
	r.Post("/comment", handlers.PostComment)
	r.Get("/admin/banned", handlers.GetAdminBanned)
	r.Post("/admin/banned", handlers.PostAdminBanned)
	r.Get(`/@{accountName:[a-zA-Z]+}`, handlers.GetAccountName)

	log.Fatal(http.ListenAndServe(":8080", r))
}
