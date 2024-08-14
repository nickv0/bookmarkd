package main

import (
	"flag"
	"fmt"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()
	if verbose {
		fmt.Println("verbose is on")
	}

	fmt.Println(flag.Args())
}
