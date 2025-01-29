package main

import (
	"fmt"
	"os"
	"path/filepath"
    "log"
)

func findExecutables(file string) ([]string, error) {
	var foundPaths []string
	path := os.Getenv("PATH")
	if path == "" {
		return nil, nil // Return nil slice and no error if PATH is not set
	}

	pathDirs := filepath.SplitList(path)

	for _, dir := range pathDirs {
		fullPath := filepath.Join(dir, file)

		// Check if the file exists and is executable
		info, err := os.Stat(fullPath)
		if err == nil {
			mode := info.Mode()
            if mode.IsRegular() && mode&0111 != 0 { // Check for regular file and executable permissions
                foundPaths = append(foundPaths, fullPath)
            }
		} else if !os.IsNotExist(err){
            // if there is some error other than file not found, return an error
            return nil, fmt.Errorf("error checking file %s: %w", fullPath, err)
        }
	}
	return foundPaths, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: which <executable> [executable2 ...]")
		os.Exit(1)
	}

	found := false
    for _, arg := range os.Args[1:] {
		foundPaths, err := findExecutables(arg)
        if err != nil {
            log.Println(err)
            os.Exit(1) // Exit on error
        }
		if len(foundPaths) > 0 {
            found = true
			for _, path := range foundPaths {
				fmt.Println(path)
			}
		}
	}
    if !found {
        os.Exit(1) // Exit with non-zero code if no executable found
    }
}