package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	// This one line turns your struct into a high-performance CLI tool
	singlechecker.Main(Analyzer)
}
