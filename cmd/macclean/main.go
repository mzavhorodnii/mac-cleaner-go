package main

import (
	"github.com/mzavhorodnii/mac-cleaner-go/internal/ui"
	"os"
)

func main() {
	root := "/Users"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	ui.StartUI(root)
}
