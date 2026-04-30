package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/experiment"
	"eml-parser/family"
	"eml-parser/formal"
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
	case "inspect":
		if len(args) != 2 {
			return usageError("inspect requires a concept name")
		}
		inspection, err := registry.Inspect(args[1])
		if err != nil {
			return err
		}
		fmt.Println(inspection.String())
		return nil
	case "search-real":
		if len(args) != 2 {
			return usageError("search-real requires a fixture name")
		}
		fixture, err := search.RealBenchmarkFixtureByName(args[1])
		if err != nil {
			return err
		}
		report, err := search.EnumerativeRealSearch(fixture, eval.Complex128Backend{}, search.SearchOptions{
			Bounds: search.Bounds{
				MaxDepth: 2,
				MaxNodes: 3,
			},
			TopN: 5,
		})
		if err != nil {
			return err
		}
		fmt.Printf("fixture: %s\n", fixture.Name)
		fmt.Println("diagnostics:")
		fmt.Println(report.Diagnostics.String())
		fmt.Println("top_candidates:")
		for i, result := range report.Results {
			fmt.Printf("%d. score=%g expr=%s\n", i+1, result.Score, result.Candidate.Normalized)
		}
		return nil
	case "formalize":
		if len(args) != 2 {
			return usageError("formalize requires a concept name")
		}
		artifact, err := formal.ExportConcept(registry, args[1])
		if err != nil {
			return err
		}
		payload, err := artifact.JSON()
		if err != nil {
			return err
		}
		fmt.Println(payload)
		return nil
	case "run-experiment":
		if len(args) != 2 {
			return usageError("run-experiment requires a spec path")
		}
		projectRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		specPath := args[1]
		if !filepath.IsAbs(specPath) {
			specPath = filepath.Join(projectRoot, specPath)
		}
		resultPath, artifact, err := experiment.RunSpecPath(projectRoot, specPath)
		if err != nil {
			return err
		}
		fmt.Printf("experiment: %s\n", artifact.ExperimentID)
		fmt.Printf("dataset: %s\n", artifact.DatasetPath)
		fmt.Printf("result: %s\n", resultPath)
		fmt.Printf("recovery_status: %s\n", artifact.RecoveryStatus)
		fmt.Println("diagnostics:")
		fmt.Println(artifact.Diagnostics.String())
		fmt.Println("top_candidates:")
		for _, candidate := range artifact.Candidates {
			fmt.Printf("%d. score=%s expr=%s\n", candidate.Rank, candidate.Score, candidate.NormalizedExpr)
		}
		return nil
	case "report-suite":
		if len(args) < 3 {
			return usageError("report-suite requires a suite id and at least one result path")
		}
		projectRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		suiteID := args[1]
		resultPaths := make([]string, 0, len(args)-2)
		for _, arg := range args[2:] {
			path := arg
			if !filepath.IsAbs(path) {
				path = filepath.Join(projectRoot, path)
			}
			resultPaths = append(resultPaths, path)
		}
		jsonPath, mdPath, summary, err := experiment.WriteSuiteReports(projectRoot, suiteID, resultPaths)
		if err != nil {
			return err
		}
		fmt.Printf("suite: %s\n", summary.SuiteID)
		fmt.Printf("json: %s\n", jsonPath)
		fmt.Printf("markdown: %s\n", mdPath)
		fmt.Printf("total_experiments: %d\n", summary.TotalExperiments)
		fmt.Printf("success_count: %d\n", summary.SuccessCount)
		fmt.Printf("failure_count: %d\n", summary.FailureCount)
		return nil
	case "gen-family-artifacts":
		projectRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		paths, artifacts, err := family.WriteCuratedArtifacts(projectRoot)
		if err != nil {
			return err
		}
		fmt.Printf("generated: %d\n", len(paths))
		for i := range paths {
			fmt.Printf("%s -> %s (%s)\n", artifacts[i].FamilyName, paths[i], artifacts[i].CanonicalKey)
		}
		return nil
	case "gen-equivalence-families":
		projectRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		paths, artifacts, err := family.WriteCuratedEquivalenceFamilies(projectRoot)
		if err != nil {
			return err
		}
		fmt.Printf("generated: %d\n", len(paths))
		for i := range paths {
			fmt.Printf("%s -> %s (members=%d)\n", artifacts[i].FamilyName, paths[i], len(artifacts[i].Members))
		}
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
	usage := "usage: emltool <list|show|deps|expand|stats|normalize|analyze|inspect|search-real|formalize|run-experiment|report-suite|gen-family-artifacts|gen-equivalence-families> [concept]"
	if prefix == "" {
		return errors.New(usage)
	}
	return fmt.Errorf("%s\n%s", prefix, usage)
}
