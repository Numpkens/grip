package main

import (
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
		Timeout: 10 * time.Second,
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
	
	staticFiles := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", staticFiles))
	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}