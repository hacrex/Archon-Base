// Command archon is the Archon Base CLI (skeleton).
package main

import (
	"fmt"
	"os"
)

const usage = `archon: CLI for Archon Base

Usage:
  archon <command> [args]

Commands:
  version     print CLI version
  init        scaffold a new project        (planned)
  dev         run the local emulator        (planned)
  db          manage vector databases       (planned)
  bucket      manage buckets                (planned)
  deploy      deploy an agent function      (planned)
  logs        stream function logs          (planned)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("archon 0.0.1-dev")
	case "init", "dev", "db", "bucket", "deploy", "logs":
		fmt.Printf("archon %s: not implemented yet\n", os.Args[1])
		os.Exit(2)
	default:
		fmt.Printf("unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}
