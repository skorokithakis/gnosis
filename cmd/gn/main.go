package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/doctrine"
	"github.com/skorokithakis/gnosis/internal/storage"
	"golang.org/x/term"
)

// printHelp prints a doctrine body followed by the shared commands
// reference, so every help surface ends with the command list.
func printHelp(body string) {
	fmt.Print(body)
	fmt.Println()
	fmt.Print(doctrine.Commands)
}

func main() {
	if len(os.Args) < 2 {
		printHelp(doctrine.Help)
		return
	}

	// Dispatch help and reject unknown commands before creating a store, so
	// that `gn help` and `gn <typo>` work even outside of a repo. Every other
	// command needs a store; it is created once after this gate instead of
	// being re-opened in each case.
	switch os.Args[1] {
	case "help":
		if len(os.Args) < 3 {
			printHelp(doctrine.Help)
			return
		}
		switch os.Args[2] {
		case "plan":
			printHelp(doctrine.Plan)
		case "review":
			printHelp(doctrine.Review)
		default:
			fmt.Fprintf(os.Stderr, "gn: unknown help topic %q\n", os.Args[2])
			os.Exit(1)
		}
		return
	case "write", "search", "latest", "show", "topics", "edit", "rm", "reindex":
		// Known command that needs a store; fall through to creation + dispatch.
	default:
		fmt.Fprintf(os.Stderr, "gn: unknown command %q\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "Run 'gn help' for a list of available commands.")
		os.Exit(1)
	}

	store, err := storage.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gn: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "write":
		if err := commands.Write(store, os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "search":
		if err := commands.Search(store, os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "latest":
		if err := commands.Latest(store, os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "gn show: missing target argument")
			os.Exit(1)
		}
		wrapWidth := 0
		if isatty.IsTerminal(os.Stdout.Fd()) {
			if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
				wrapWidth = min(width, 80)
			}
		}
		if err := commands.Show(store, os.Args[2], wrapWidth, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn show: %v\n", err)
			os.Exit(1)
		}
	case "topics":
		if err := commands.Topics(store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "edit":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "gn: edit: missing entry ID")
			os.Exit(1)
		}
		if len(os.Args) > 4 {
			fmt.Fprintln(os.Stderr, "usage: gn edit <id> [text]")
			os.Exit(1)
		}
		if err := commands.Edit(store, os.Args[2], os.Args[3:], os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "rm":
		if err := commands.Remove(store, os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	case "reindex":
		// repoRoot is resolved separately rather than recovered from the store
		// because the store does not expose the root it was built from.
		repoRoot, err := storage.FindRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
		if err := commands.Reindex(repoRoot, store, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "gn: %v\n", err)
			os.Exit(1)
		}
	}
}
