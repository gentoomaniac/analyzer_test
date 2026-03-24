package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	// analysistest.TestData() figures out the absolute path to your 'testdata' folder
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, Analyzer, "mypkg")
}
