package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"eml-parser/concepts"
	"eml-parser/normalize"
	"eml-parser/search"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError("")
	}

	registry := concepts.StandardLibrary()
	switch args[0] {
	case "list":
		for _, name := range registry.Names() {
			fmt.Println(name)
		}
		return nil
	case "show":
		if len(args) != 2 {
			return usageError("show requires a concept name")
		}
		def, ok := registry.Definition(args[1])
		if !ok {
			return fmt.Errorf("unknown concept %q", args[1])
		}
		if len(def.Params) == 0 {
			fmt.Printf("%s := %s\n", def.Name, def.Body)
			return nil
		}
		fmt.Printf("%s(%s) := %s\n", def.Name, strings.Join(def.Params, ", "), def.Body)
		return nil
	case "deps":
		if len(args) != 2 {
			return usageError("deps requires a concept name")
		}
		direct, err := registry.DirectDependencies(args[1])
		if err != nil {
			return err
		}
		transitive, err := registry.TransitiveDependencies(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("direct: %s\n", joinOrNone(direct))
		fmt.Printf("transitive: %s\n", joinOrNone(transitive))
		return nil
	case "expand":
		if len(args) != 2 {
			return usageError("expand requires a concept name")
		}
		expr, err := registry.ExpandSymbolic(args[1])
		if err != nil {
			return err
		}
		fmt.Println(expr.String())
		return nil
	case "stats":
		if len(args) != 2 {
			return usageError("stats requires a concept name")
		}
		stats, err := registry.Stats(args[1])
		if err != nil {
			return err
		}
		fmt.Println(stats.String())
		return nil
	case "normalize":
		if len(args) != 2 {
			return usageError("normalize requires a concept name")
		}
		expr, err := registry.ExpandSymbolic(args[1])
		if err != nil {
			return err
		}
		fmt.Println(normalize.Expr(expr).String())
		return nil
	case "analyze":
		if len(args) != 2 {
			return usageError("analyze requires a concept name")
		}
		expr, err := registry.ExpandSymbolic(args[1])
		if err != nil {
			return err
		}
		candidate := search.NewCandidate(expr)
		fmt.Printf("concept: %s\n", args[1])
		fmt.Printf("key: %s\n", candidate.Key)
		fmt.Printf("nodes: %d\n", candidate.Stats.NodeCount)
		fmt.Printf("depth: %d\n", candidate.Stats.TreeDepth)
		fmt.Printf("leaves: %d\n", candidate.Stats.LeafCount)
		fmt.Printf("normalized: %s\n", candidate.Normalized)
		return nil
	default:
		return usageError(fmt.Sprintf("unknown command %q", args[0]))
	}
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func usageError(prefix string) error {
	usage := "usage: emltool <list|show|deps|expand|stats|normalize|analyze> [concept]"
	if prefix == "" {
		return errors.New(usage)
	}
	return fmt.Errorf("%s\n%s", prefix, usage)
}
