package main

import (
	"log"
	"os"

	"github.com/fzerorubigd/cardx/cards"
)

func main() {
	var (
		root string
	)

	left := os.Args
	if len(left) != 2 {
		log.Fatalf("no project root provided: %s /path/to/project/root", os.Args[0])
	}
	root = left[1]

	p, err := cards.NewProject(root)
	if err != nil {
		log.Fatal(err)
	}

	if err := cards.Generate(p, p.Output, AssetNames(), MustAsset); err != nil {
		log.Fatal(err)
	}
}
