package handlers

import (
	"log"
	"net/http"

	"github.com/catatsuy/private-isu/webapp/golang/models"
	"github.com/catatsuy/private-isu/webapp/golang/utils"
	"golang.org/x/sync/errgroup"
)

func GetAdminBanned(w http.ResponseWriter, r *http.Request) {
	me := utils.GetSessionUser(r)
	if !utils.IsLogin(me) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if me.Authority == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	users := []models.User{}
	err := db.Select(&users, "SELECT * FROM `users` WHERE `authority` = 0 AND `del_flg` = 0 ORDER BY `created_at` DESC")
	if err != nil {
		log.Print(err)
		return
	}

	template.Must(template.ParseFiles(
		getTemplPath("layout.html"),
		getTemplPath("banned.html")),
	).Execute(w, struct {
		Users     []models.User
		Me        models.User
		CSRFToken string
	}{users, me, utils.GetCSRFToken(r)})
}

func PostAdminBanned(w http.ResponseWriter, r *http.Request) {
	me := utils.GetSessionUser(r)
	if !utils.IsLogin(me) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if me.Authority == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.FormValue("csrf_token") != utils.GetCSRFToken(r) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Print(err)
		return
	}

	// Process bans in parallel
	eg := errgroup.Group{}
	for _, id := range r.Form["uid[]"] {
		id := id // capture loop variable
		eg.Go(func() error {
			_, err := db.Exec("UPDATE `users` SET `del_flg` = ? WHERE `id` = ?", 1, id)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		log.Print(err)
		return
	}

	http.Redirect(w, r, "/admin/banned", http.StatusFound)
} 