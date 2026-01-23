// Package handlers implements the HTTP interface for the GRIP aggregator.
//It serves the dual purpose of providing both web and api.
package handlers

import (
	"encoding/json"
	"github.com/Numpkens/grip/internal/logic"
	"html/template"
	"net/http"
	"time"
	"log"
)

// Handler maintains the dependencies required to serve GRIP requests.
type Handler struct {
	Templ  *template.Template
	Engine *logic.Engine
}
// TemplateData sends server performance information for the template to consume and display
type TemplateData struct {
	Results []logic.Post
	Query   string
	Latency string
}

// HandleHome aggregates and serves blog posts via HTML or JSON.
// @Summary      Search Aggregated Blogs
// @Description  Returns the top 20 newest posts. Detects 'Accept' header for JSON/HTML toggle.
// @Produce      json
// @Produce      html
// @Param        q    query     string  false  "Search Keyword (defaults to 'golang')"
// @Success      200  {array}   logic.Post "Successfully retrieved posts"
// @Failure      400  {string}  string     "Bad Request: Invalid query parameters"
// @Failure      500  {string}  string     "Template Execution Error: Template or JSON encoding failed"
// @Failure      504  {string}  string     "Gateway Timeout: All sources failed to respond within 2s"
// @Router       / [get]
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "golang"
	}

	start := time.Now()
	posts := h.Engine.Collect(r.Context(), query)
	latency := time.Since(start).Truncate(time.Millisecond).String()

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
		return
	}

	data := TemplateData{
		Results: posts,
		Query:   query,
		Latency: latency,
	}

	err := h.Templ.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}
}