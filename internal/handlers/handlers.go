package handlers

import (
	"github.com/Numpkens/grip/internal/logic"
	"html/template"
	"net/http"
	"encoding/json"
)

type Handler struct {
	Templ  *template.Template
	Engine *logic.Engine
}
//HandleHome godoc
//@Summary Search Aggregated Blogs for Deveolpers
//@Description Receives the 20 most recent posts from multiple sources(Dev.to, Hashnode, Hackernews etc...)
//@Produce json
//@Produce html
//@Param q query string false "Search Keyword(defaults to golang)"
//@Success 200 {array} logic.Post
//@Failure 500 {string} string "Internal Server Error"
//@Router / [get]
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "golang"
	}

	posts := h.Engine.Collect(r.Context(), query)

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
		return
	}

	err := h.Templ.Execute(w, posts)
	if err != nil {
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
	}
}
