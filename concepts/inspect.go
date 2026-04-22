package concepts

import (
	"fmt"
	"strings"

	"eml-parser/ast"
	"eml-parser/normalize"
)

// Inspection summarizes a concept across the concept layer and raw EML layer.
type Inspection struct {
	Name               string
	Definition         string
	Expanded           ast.Expr
	Normalized         ast.Expr
	ExpandedStats      ExpandedStats
	NormalizedStats    ExpandedStats
	DependencyPaths    [][]string
	ExpandedNodeDelta  int
	ExpandedDepthDelta int
}

// Inspect returns a combined concept/raw-tree view suitable for tooling.
func (r *Registry) Inspect(name string) (Inspection, error) {
	def, ok := r.Definition(name)
	if !ok {
		return Inspection{}, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}

	expanded, err := r.ExpandSymbolic(name)
	if err != nil {
		return Inspection{}, err
	}
	normalized := normalize.Expr(cloneAST(expanded))
	expandedStats := StatsForExpr(expanded)
	normalizedStats := StatsForExpr(normalized)
	paths, err := r.DependencyPaths(name)
	if err != nil {
		return Inspection{}, err
	}

	definition := def.Name
	if len(def.Params) > 0 {
		definition += "(" + strings.Join(def.Params, ", ") + ")"
	}
	definition += " := " + def.Body.String()

	return Inspection{
		Name:               name,
		Definition:         definition,
		Expanded:           expanded,
		Normalized:         normalized,
		ExpandedStats:      expandedStats,
		NormalizedStats:    normalizedStats,
		DependencyPaths:    paths,
		ExpandedNodeDelta:  expandedStats.NodeCount - normalizedStats.NodeCount,
		ExpandedDepthDelta: expandedStats.TreeDepth - normalizedStats.TreeDepth,
	}, nil
}

func (i Inspection) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "concept: %s\n", i.Name)
	fmt.Fprintf(&b, "definition: %s\n", i.Definition)
	fmt.Fprintf(&b, "expanded: %s\n", i.Expanded)
	fmt.Fprintf(&b, "normalized: %s\n", i.Normalized)
	fmt.Fprintf(&b, "expanded_nodes: %d\n", i.ExpandedStats.NodeCount)
	fmt.Fprintf(&b, "normalized_nodes: %d\n", i.NormalizedStats.NodeCount)
	fmt.Fprintf(&b, "node_delta: %d\n", i.ExpandedNodeDelta)
	fmt.Fprintf(&b, "expanded_depth: %d\n", i.ExpandedStats.TreeDepth)
	fmt.Fprintf(&b, "normalized_depth: %d\n", i.NormalizedStats.TreeDepth)
	fmt.Fprintf(&b, "depth_delta: %d\n", i.ExpandedDepthDelta)
	fmt.Fprintf(&b, "expanded_leaves: %d\n", i.ExpandedStats.LeafCount)
	fmt.Fprintf(&b, "normalized_leaves: %d\n", i.NormalizedStats.LeafCount)
	fmt.Fprintf(&b, "dependency_paths:\n")
	if len(i.DependencyPaths) == 0 {
		fmt.Fprintf(&b, "- (none)\n")
		return strings.TrimRight(b.String(), "\n")
	}
	for _, path := range i.DependencyPaths {
		fmt.Fprintf(&b, "- %s\n", strings.Join(path, " -> "))
	}
	return strings.TrimRight(b.String(), "\n")
}

// DependencyPaths returns root-to-leaf concept call chains for a named concept.
func (r *Registry) DependencyPaths(name string) ([][]string, error) {
	if _, ok := r.defs[name]; !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}
	paths, err := r.collectDependencyPaths(name, nil)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func (r *Registry) collectDependencyPaths(name string, stack []string) ([][]string, error) {
	for _, existing := range stack {
		if existing == name {
			return nil, fmt.Errorf("%w: %v -> %s", ErrConceptCycle, stack, name)
		}
	}
	def := r.defs[name]
	direct := sortedKeys(collectDirectDependencies(def.Body))
	if len(direct) == 0 {
		return [][]string{{name}}, nil
	}

	var out [][]string
	nextStack := append(stack, name)
	for _, dep := range direct {
		if _, ok := r.defs[dep]; !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, dep)
		}
		subpaths, err := r.collectDependencyPaths(dep, nextStack)
		if err != nil {
			return nil, err
		}
		for _, subpath := range subpaths {
			path := make([]string, 0, len(subpath)+1)
			path = append(path, name)
			path = append(path, subpath...)
			out = append(out, path)
		}
	}
	return out, nil
}
