package handlers

import (
	"html/template"
	"net/http"
	"github.com/Numpkens/grip/internal/logic"
)

type Handler struct {
	Templ *template.Template
	Engine *logic.Engine
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {

	posts := h.Engine.FetchAll("golang")

	err := h.Templ.Execute(w, posts)
	if err != nil {
		http.Error(w, "Internal server Error", http.StatusInternalServerError)
	}
}


