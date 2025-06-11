package utils

import (
	"net/http"

	"github.com/catatsuy/private-isu/webapp/golang/cache"
	"github.com/catatsuy/private-isu/webapp/golang/models"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"log"
)

var (
	db    *sqlx.DB
	store sessions.Store
)

func InitSession(database *sqlx.DB, sessionStore sessions.Store) {
	db = database
	store = sessionStore
}

func GetSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "isuconp-go.session")
	return session
}

func GetSessionUser(r *http.Request) models.User {
	session := GetSession(r)
	uid, ok := session.Values["user_id"]
	if !ok || uid == nil {
		return models.User{}
	}

	// Try to get from cache first
	userID := uid.(int)
	cachedUser, err := cache.GetUserFromCache(userID)
	if err != nil {
		// Log error but continue to get from DB
		log.Printf("Failed to get user from cache: %v", err)
	} else if cachedUser != nil {
		return *cachedUser
	}

	// Cache miss or error, get from DB
	u := models.User{}
	err = db.Get(&u, "SELECT * FROM `users` WHERE `id` = ?", userID)
	if err != nil {
		return models.User{}
	}

	// Set cache for next time
	if err := cache.SetUserCache(&u); err != nil {
		log.Printf("Failed to set user cache: %v", err)
	}

	return u
}

func GetFlash(w http.ResponseWriter, r *http.Request, key string) string {
	session := GetSession(r)
	value, ok := session.Values[key]

	if !ok || value == nil {
		return ""
	}

	delete(session.Values, key)
	session.Save(r, w)
	return value.(string)
}

func IsLogin(u models.User) bool {
	return u.ID != 0
}

func GetCSRFToken(r *http.Request) string {
	session := GetSession(r)
	csrfToken, ok := session.Values["csrf_token"]
	if !ok {
		return ""
	}
	return csrfToken.(string)
} 