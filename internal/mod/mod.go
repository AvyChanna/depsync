package mod

import (
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/AvyChanna/depsync/internal/set"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

const (
	depSizeHint = 100
	modSizeHint = 5
)

type DepMap = map[string][]ModLine

type ModLine struct {
	ModFile string
	Version string
}

// FindAll scans the provided base paths for Go mod and work files,
// and returns a slice of paths to the discovered go.mod files.
func FindAll(searchPaths []string) []string {
	mods := make(set.Set[string], 2*len(searchPaths))
	for _, basePath := range searchPaths {
		basePath, err := filepath.Abs(basePath)
		panicOnErr(err)
		baseName := filepath.Base(basePath)

		
		if baseName == "go.mod" {
			if checkFileExists(basePath) {
				mods.Insert(basePath)
			}
		} else if baseName == "go.work" {
			if checkFileExists(basePath) {
				mods.InsertMany(getModsFromWork(basePath))
			}
		} else {
			goModPath := filepath.Join(basePath, "go.mod")
			if checkFileExists(goModPath) {
				mods.Insert(goModPath)
			}

			goWorkPath := filepath.Join(basePath, "go.work")
			if checkFileExists(goWorkPath) {
				mods.InsertMany(getModsFromWork(goWorkPath))
			}
		}
	}
	return slices.Collect(maps.Keys(mods))
}

// ParseMods reads and parses a list of go.mod files.
func ParseMods(modFilePaths []string) DepMap {
	res := make(DepMap, depSizeHint)
	for _, modPath := range modFilePaths {
		data, err := os.ReadFile(modPath)
		panicOnErr(err)

		mod, err := modfile.Parse(modPath, data, nil)
		panicOnErr(err)

		for _, req := range mod.Require {
			dep := req.Mod.Path
			version := req.Mod.Version
			if res[dep] == nil {
				res[dep] = make([]ModLine, 0, modSizeHint)
			}
			res[dep] = append(res[dep], ModLine{ModFile: modPath, Version: version})
		}
	}

	for i := range res {
		slices.SortFunc(res[i], sortVersions)
	}
	return res
}

// getModsFromWork parses a go.work file to extract the mod paths.
func getModsFromWork(workPath string) []string {
	basePath := filepath.Dir(workPath)
	res := make([]string, 0, modSizeHint)
	data, err := os.ReadFile(workPath)
	panicOnErr(err)

	work, err := modfile.ParseWork(workPath, data, nil)
	panicOnErr(err)

	for _, use := range work.Use {
		innerModPath := filepath.Join(basePath, use.Path, "go.mod")
		if !checkFileExists(innerModPath) {
			continue
		}
		res = append(res, innerModPath)
	}
	return res
}

// sortVersions sorts ModLine by semver version and then by mod file path.
func sortVersions(a, b ModLine) int {
	verCmp := semver.Compare(a.Version, b.Version)
	if verCmp != 0 {
		return verCmp
	}

	if a.ModFile < b.ModFile {
		return -1
	}

	return 1
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
