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

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "golang"
	}

	posts := h.Engine.FetchAll(query)

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
