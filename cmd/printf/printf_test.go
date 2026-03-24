package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	// Import your analyzer package
)

func TestAnalyzer(t *testing.T) {
	// analysistest.TestData() figures out the absolute path to your 'testdata' folder
	testdata := analysistest.TestData()

	// analysistest.Run takes:
	// 1. The testing object
	// 2. The path to the testdata directory
	// 3. Your Analyzer struct
	// 4. The name of the fake package inside testdata/src/ to analyze (we used "a")
	analysistest.Run(t, testdata, Analyzer, "mypkg")
}
