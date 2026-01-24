package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
	"github.com/Numpkens/grip/internal/logic"
	"github.com/Numpkens/grip/internal/logic/sources"
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

	fmt.Printf("Searching for %s...\n", query)
	posts := engine.Collect(context.Background(), query)

	if len(posts) == 0 {
		fmt.Println("No results found.")
		return
	}


	for i, p := range posts {
		fmt.Printf("[%d] %-60s | %s\n", i+1, p.Title, p.Source)
	}


	fmt.Print("\nEnter number to open (0 to exit): ")
	var choice int
	fmt.Scanln(&choice)

	if choice > 0 && choice <= len(posts) {
		openURL(posts[choice-1].URL)
	}
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: 
		cmd = "xdg-open"
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}