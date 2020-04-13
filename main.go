package main

import (
	"log"
	"os"

	"github.com/fzerorubigd/cardx/cards"
	"github.com/ogier/pflag"
)

func main() {
	var (
		root   string
		output string
	)

	pflag.StringVar(&output, "output", "", "Output to create the generated folder (default to project folder/generated) ")

	pflag.Parse()

	left := pflag.Args()
	if len(left) != 1 {
		log.Fatalf("no project root provided: %s [OPTIONS] /path/to/project/root", os.Args[0])
	}
	root = left[0]

	p, err := cards.NewProject(root)
	if err != nil {
		log.Fatal(err)
	}

	if err := cards.Generate(p, output, AssetNames(), MustAsset); err != nil {
		log.Fatal(err)
	}

}
