package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version":
		fmt.Printf("MPM (Model Package Manager) v%s\n", version)

	case "list":
		handleList(os.Args[2:])

	case "detect":
		handleDetect(os.Args[2:])

	case "recommend":
		handleRecommend(os.Args[2:])

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: mpm search <query>")
			os.Exit(1)
		}
		handleSearch(os.Args[2:])

	case "install":
		handleInstall(os.Args[2:])

	case "active":
		handleActive(os.Args[2:])

	case "set-active":
		handleSetActive(os.Args[2:])

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`MPM - Model Package Manager

Usage: mpm <command> [options]

Commands:
  list                List all available models
  detect              Detect system hardware
  recommend           Recommend a model based on hardware
  search <query>      Search models by name/capability
  install <model>     Install a model
  active              Show currently active model
  set-active <model>  Set the active model
  version             Show version
  help                Show this help

Options:
  --json              Output as JSON
  --silent            Suppress output
  --family <name>     Filter by model family
  --tag <tag>         Filter by tag
  --capability <cap>  Filter by capability
  --size <size>       Filter by size (0.6b, 1.7b, 4b, 8b)
`)
}
