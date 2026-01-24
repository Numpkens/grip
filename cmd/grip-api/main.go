package main

import (
	"net/http"
	"time"
	"github.com/Numpkens/grip/internal/logic"         
	"github.com/Numpkens/grip/internal/logic/sources"
	"github.com/Numpkens/grip/internal/handlers"
	_ "github.com/Numpkens/grip/docs"
	httpSwagger "github.com/swaggo/http-swagger" 
)

// @title           GRIP API
// @version         1.0
// @description     A high-performance developer blog aggregator proxy.
// @host            localhost:8080
// @BasePath        /
func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	
	engine := logic.NewEngine([]logic.Source{
		&sources.DevTo{Client: client},
		&sources.HackerNews{Client: client},
		&sources.Hashnode{Client: client},
		&sources.BootDev{Client: client},
		&sources.Lobsters{Client: client, BaseURL: "https://lobste.rs"},
		&sources.FreeCodeCamp{Client: client, BaseURL: "https://www.freecodecamp.org"},
	})

	h := &handlers.Handler{
		Engine: engine,
	}

	http.Handle("/swagger/", httpSwagger.WrapHandler)
	http.HandleFunc("/", h.HandleHome)

	http.ListenAndServe(":8080", nil)
}