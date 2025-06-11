package utils

import (
	"net/http"

	"github.com/catatsuy/private-isu/webapp/golang/models"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
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

	u := models.User{}
	err := db.Get(&u, "SELECT * FROM `users` WHERE `id` = ?", uid)
	if err != nil {
		return models.User{}
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