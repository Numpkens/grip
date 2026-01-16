package main

import (
	"html/template"
	"log"
	"net/http"
	"github.com/Numpkens/grip/internal/handlers"
	"github.com/Numpkens/grip/internal/logic"
	"github.com/swaggo/http-swagger"
  _ "github.com/Numpkens/grip/docs"
)
// @title           GRIP API
// @version         1.0
// @description     Concurrent Developer News Aggregator.
// @host            localhost:8080
// @BasePath        /
// @query.collection.format multi
func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	engine := &logic.Engine{
		Sources: []logic.Source{
			&logic.DevTo{},
			&logic.HackerNews{},
			&logic.Hashnode{},
			&logic.BootDev{},
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
