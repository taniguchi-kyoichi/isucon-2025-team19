package handlers

import (
	"log"
	"net/http"
	"regexp"

	"github.com/catatsuy/private-isu/webapp/golang/cache"
	"github.com/catatsuy/private-isu/webapp/golang/models"
	"github.com/catatsuy/private-isu/webapp/golang/utils"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

var (
	db    *sqlx.DB
	store sessions.Store
)

func InitHandlers(database *sqlx.DB, sessionStore sessions.Store) {
	db = database
	store = sessionStore
}

func validateUser(accountName, password string) bool {
	return regexp.MustCompile(`\A[0-9a-zA-Z_]{3,}\z`).MatchString(accountName) &&
		regexp.MustCompile(`\A[0-9a-zA-Z_]{6,}\z`).MatchString(password)
}

func GetLogin(w http.ResponseWriter, r *http.Request) {
	me := utils.GetSessionUser(r)

	if utils.IsLogin(me) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	template.Must(template.ParseFiles(
		getTemplPath("layout.html"),
		getTemplPath("login.html")),
	).Execute(w, struct {
		Me    models.User
		Flash string
	}{me, utils.GetFlash(w, r, "notice")})
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	if utils.IsLogin(utils.GetSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	u := tryLogin(r.FormValue("account_name"), r.FormValue("password"))

	if u != nil {
		session := utils.GetSession(r)
		session.Values["user_id"] = u.ID
		session.Values["csrf_token"] = utils.SecureRandomStr(16)
		session.Save(r, w)

		// Cache the user data
		if err := cache.SetUserCache(u); err != nil {
			log.Printf("Failed to set user cache on login: %v", err)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		session := utils.GetSession(r)
		session.Values["notice"] = "アカウント名かパスワードが間違っています"
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func GetRegister(w http.ResponseWriter, r *http.Request) {
	if utils.IsLogin(utils.GetSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	template.Must(template.ParseFiles(
		getTemplPath("layout.html"),
		getTemplPath("register.html")),
	).Execute(w, struct {
		Me    models.User
		Flash string
	}{models.User{}, utils.GetFlash(w, r, "notice")})
}

func PostRegister(w http.ResponseWriter, r *http.Request) {
	if utils.IsLogin(utils.GetSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	accountName, password := r.FormValue("account_name"), r.FormValue("password")

	validated := validateUser(accountName, password)
	if !validated {
		session := utils.GetSession(r)
		session.Values["notice"] = "アカウント名は3文字以上、パスワードは6文字以上である必要があります"
		session.Save(r, w)

		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	exists := 0
	err := db.Get(&exists, "SELECT 1 FROM users WHERE `account_name` = ?", accountName)
	if err == nil && exists == 1 {
		session := utils.GetSession(r)
		session.Values["notice"] = "アカウント名がすでに使われています"
		session.Save(r, w)

		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	query := "INSERT INTO `users` (`account_name`, `passhash`) VALUES (?,?)"
	result, err := db.Exec(query, accountName, utils.CalculatePasshash(accountName, password))
	if err != nil {
		log.Print(err)
		return
	}

	session := utils.GetSession(r)
	uid, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		return
	}

	// Get the newly created user
	newUser := models.User{}
	err = db.Get(&newUser, "SELECT * FROM `users` WHERE `id` = ?", uid)
	if err != nil {
		log.Print(err)
		return
	}

	// Cache the new user
	if err := cache.SetUserCache(&newUser); err != nil {
		log.Printf("Failed to set user cache on register: %v", err)
	}

	session.Values["user_id"] = uid
	session.Values["csrf_token"] = utils.SecureRandomStr(16)
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func GetLogout(w http.ResponseWriter, r *http.Request) {
	session := utils.GetSession(r)
	
	// Delete user cache if logged in
	if uid, ok := session.Values["user_id"]; ok && uid != nil {
		if err := cache.DeleteUserCache(uid.(int)); err != nil {
			log.Printf("Failed to delete user cache on logout: %v", err)
		}
	}

	delete(session.Values, "user_id")
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
} 