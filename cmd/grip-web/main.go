package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/Numpkens/grip/docs"
	"github.com/Numpkens/grip/internal/handlers"
	"github.com/Numpkens/grip/internal/logic"
	"github.com/Numpkens/grip/internal/logic/sources"
	"github.com/swaggo/http-swagger"
)

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	engine := &logic.Engine{
		Sources: []logic.Source{
			&sources.DevTo{Client: httpClient},
			&sources.HackerNews{Client: httpClient},
			&sources.Hashnode{Client: httpClient},
			&sources.BootDev{Client: httpClient},
			&sources.Lobsters{Client: httpClient, BaseURL: "https://lobste.rs"},
			&sources.FreeCodeCamp{Client: httpClient, BaseURL: "https://www.freecodecamp.org"},
		},
	}

	h := &handlers.Handler{
		Templ:  tmpl,
		Engine: engine,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HandleHome)

	// @Summary Search posts
	// @Description Returns raw search results as JSON
	// @Tags search
	// @Produce json
	// @Param q query string false "Search query"
	// @Success 200 {array} logic.Post
	// @Router /api/search [get]
	// @Summary Search posts
	// @Description Returns raw search results as JSON
	// @Tags search
	// @Produce json
	// @Param q query string false "Search query"
	// @Success 200 {array} logic.Post
	// @Router /api/search [get]
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		if query == "" {
			query = "golang"
		}

		posts := engine.Collect(r.Context(), query) 

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if err := json.NewEncoder(w).Encode(posts); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	staticFiles := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", staticFiles))
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("GRIP starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}