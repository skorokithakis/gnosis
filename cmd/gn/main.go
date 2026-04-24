package main

import (
	"fmt"
	"os"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/doctrine"
	"github.com/skorokithakis/gnosis/internal/storage"
)

func notImplemented(command string) {
	fmt.Fprintf(os.Stderr, "gn: %s: not implemented\n", command)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(doctrine.Help)
		return
	}

	switch os.Args[1] {
	case "help":
		if len(os.Args) < 3 {
			fmt.Print(doctrine.Help)
			return
		}
		switch os.Args[2] {
		case "review":
			fmt.Print(doctrine.Review)
		default:
			fmt.Fprintf(os.Stderr, "gn: unknown help topic %q\n", os.Args[2])
			os.Exit(1)
		}
	case "write":
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Write(store, os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "search":
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Search(store, os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "gn show: missing target argument")
			os.Exit(1)
		}
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Show(store, os.Args[2], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn show: %v\n", err)
			os.Exit(1)
		}
	case "topics":
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Topics(store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "edit":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "gn: edit: missing entry ID")
			os.Exit(1)
		}
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Edit(store, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "rm":
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Remove(store, os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "reindex":
		repoRoot, err := storage.FindRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Reindex(repoRoot, store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "gn: unknown command %q\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "Run 'gn help' for a list of available commands.")
		os.Exit(1)
	}
}
