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
// HandleHome godoc
// @Summary      Search Aggregated Blogs
// @Description  Aggregates the top 20 newest posts from Dev.to, Hashnode, HN, etc.
// @Description  Supports a 2-second timeout and sorts by date using a Min-Heap.
// @Produce      json
// @Produce      html
// @Param        q    query     string  false  "Search Keyword (defaults to 'golang')"
// @Success      200  {array}   logic.Post "Returns JSON if Accept header is application/json"
// @Success      200  {string}  string     "Returns HTML Card View by default"
// @Failure      500  {string}  string     "Internal Server Error"
// @Router       / [get]
//
// HandleHome serves as the main search interface. It detects the accept header to decide whether to use JSON or HTML.
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
