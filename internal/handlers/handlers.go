package handlers

import (
	"github.com/Numpkens/grip/internal/logic"
	"html/template"
	"net/http"
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

	err := h.Templ.Execute(w, posts)
	if err != nil {
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
	}
}
