package logic_test

import (
	"context"
	"fmt"
	"github.com/Numpkens/grip/internal/logic"
)

func ExampleEngine_Collect() {
	
	engine := &logic.Engine{
		Sources: []logic.Source{},
	}
	
	posts := engine.Collect(context.Background(), "golang")
	fmt.Println(len(posts))
	
}