package main

import (
	"github.com/Ghvstcode/sleeng/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
