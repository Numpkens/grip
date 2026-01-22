package logic_test

import (
	"context"
	"fmt"
	"github.com/Numpkens/grip/internal/logic"
)

func ExampleEngine_Collect() {
	engine := logic.NewEngine([]logic.Source{ /* mock sources */ })
	posts := engine.Collect(context.Background(), "golang")
	fmt.Println(len(posts))
	// Output: 20
}