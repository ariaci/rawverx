package main

import (
	"fmt"
	"log"

	"github.com/ariaci/rawverx/examples/complete-workflow/version"
)

func main() {
	info, err := version.NewInfo()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(info)
}
