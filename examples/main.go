package main

import (
	"fmt"

	"github.com/oarkflow/fastac"
)

func main() {
	e, err := fastac.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(e.Enforce("alice", map[string]any{"price": 28, "brand": "puma"}, "read"))
}
