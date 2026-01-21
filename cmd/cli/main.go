package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"github.com/Numpkens/grip/internal/logic"
	"github.com/Numpkens/grip/internal/logic/sources"
	"os"
)

func main() {

	query := "golang"
	if len(os.Args) > 1 {
		query = os.Args[1]
	}
	client := &http.Client{Timeout: 5 * time.Second}

	engine := logic.NewEngine([]logic.Source{
		&sources.DevTo{Client: client},
			&sources.HackerNews{Client: client},
			&sources.Hashnode{Client: client},
			&sources.BootDev{Client: client},
			&sources.Lobsters{Client: client, BaseURL: "https://lobste.rs"},
			&sources.FreeCodeCamp{Client: client, BaseURL: "https://www.freecodecamp.org"},
	})

	fmt.Printf("Searching for %s across %d sources...\n", query, len(engine.Sources))
	posts := engine.Collect(context.Background(), query)

	if len(posts) == 0 {
		fmt.Println("No results found. Check network or API rate limits.")
		return
	}

	for i, p := range posts {
		fmt.Printf("[%d] %s (%s)\n", i+1, p.Title, p.Source)
	}
}