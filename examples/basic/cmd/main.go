package main

import (
	"fmt"
	"log"

	"github.com/gopherust-io/env/examples/basic"
)

func main() {
	cfg, err := basic.Load()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("config: %+v\n", cfg.Masked())
}
