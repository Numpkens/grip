package handlers

import (
	"html/template"
	"net/http"

	"github.com/Numpkens/grip/internal/logic"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	post := logic.GetMockPosts()

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, post)
}
