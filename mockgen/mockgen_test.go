package main

import (
	"testing"
)

func TestMockgen(t *testing.T) {
	run(runParam{
		sourceDir:      "./testdata",
		outputFile:     "src_mock.go",
		mockInterfaces: "!ServiceV2",
	})
}
