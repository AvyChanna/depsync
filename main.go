package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AvyChanna/depsync/internal/mod"
)

const usageStr = `Usage: %s <path> [<path> ...]
	path = Path to go.mod/go.work or their parent directory
`

func main() {
	searchPaths := os.Args[1:]
	if len(searchPaths) == 0 {
		fmt.Printf(usageStr, filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	allMods := mod.FindAll(searchPaths)
	allDeps := mod.ParseMods(allMods)
	hasMismatch := check(allDeps)

	if hasMismatch {
		os.Exit(1)
	}
}

// check checks for mismatched versions in deps.
// It prints the mismatches and returns true if any are found.
func check(allDeps mod.DepMap) bool {
	hasMismatch := false
	for dep, modVer := range allDeps {
		versions := modLinesToVersions(modVer)
		uniqVersions := uniqSortedVals(versions)

		if len(uniqVersions) == 1 {
			continue
		}

		hasMismatch = true
		fmt.Printf("[!] Mismatch for `%s`. Found %d versions - {%s}\n", dep, len(uniqVersions), strings.Join(uniqVersions, ","))

		for _, mv := range modVer {
			fmt.Printf("\t- %s: %s\n", mv.Version, mv.ModFile)
		}
	}
	return hasMismatch
}

// uniqSortedVals returns a sorted slice of unique values from the input slice.
// The input slice must be sorted for this function to work correctly.
func uniqSortedVals[T comparable](input []T) []T {
	unique := []T{input[0]}
	for i := 1; i < len(input); i++ {
		if input[i] != input[i-1] {
			unique = append(unique, input[i])
		}
	}
	return unique
}

func modLinesToVersions(modVer []mod.ModLine) []string {
	res := make([]string, len(modVer))
	for i, mv := range modVer {
		res[i] = mv.Version
	}
	return res
}
