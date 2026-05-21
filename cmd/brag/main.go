package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "add":
		handleAdd()
	case "update":
		handleUpdate()
	case "enrich":
		handleEnrich()
	case "list":
		handleList()
	case "report":
		handleReport()
	case "review":
		handleReview()
	case "remove":
		handleRemove()
	case "clear":
		handleClear()
	case "export":
		handleExport()
	case "config":
		handleConfig()
	case "help":
		handleHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'brag help' for usage information.")
		os.Exit(1)
	}
}
