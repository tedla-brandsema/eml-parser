package formal

import (
	"encoding/json"
	"fmt"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/normalize"
	"eml-parser/search"
)

// Artifact is a proof-friendly export of normalized raw EML.
type Artifact struct {
	FormatVersion string      `json:"format_version"`
	RootID        int         `json:"root_id"`
	Nodes         []Node      `json:"nodes"`
	Expression    string      `json:"expression"`
	Provenance    *Provenance `json:"provenance,omitempty"`
}

// Node is one normalized raw EML node in a deterministic exported form.
type Node struct {
	ID      int    `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name,omitempty"`
	LeftID  int    `json:"left_id,omitempty"`
	RightID int    `json:"right_id,omitempty"`
}

// Provenance records how an exported raw tree was obtained.
type Provenance struct {
	Source       string             `json:"source"`
	Concept      *ConceptProvenance `json:"concept,omitempty"`
	CandidateKey string             `json:"candidate_key,omitempty"`
}

// ConceptProvenance records concept-layer information retained during export.
type ConceptProvenance struct {
	Name            string     `json:"name"`
	Definition      string     `json:"definition"`
	DependencyPaths [][]string `json:"dependency_paths,omitempty"`
}

// ExportExpr normalizes a raw EML expression and exports it as a proof-friendly
// intermediate artifact.
func ExportExpr(expr ast.Expr) Artifact {
	normalized := normalize.Expr(clone(expr))
	rootID, nodes := exportNodes(normalized)
	return Artifact{
		FormatVersion: "eml-formal-v1",
		RootID:        rootID,
		Nodes:         nodes,
		Expression:    normalized.String(),
	}
}

// ExportCandidate exports a normalized search candidate with candidate
// provenance retained.
func ExportCandidate(candidate search.Candidate) Artifact {
	rootID, nodes := exportNodes(candidate.Normalized)
	return Artifact{
		FormatVersion: "eml-formal-v1",
		RootID:        rootID,
		Nodes:         nodes,
		Expression:    candidate.Normalized.String(),
		Provenance: &Provenance{
			Source:       "candidate",
			CandidateKey: candidate.Key,
		},
	}
}

// ExportConcept exports a named concept after expansion and normalization,
// retaining concept provenance where available.
func ExportConcept(registry *concepts.Registry, name string) (Artifact, error) {
	inspection, err := registry.Inspect(name)
	if err != nil {
		return Artifact{}, err
	}
	rootID, nodes := exportNodes(inspection.Normalized)
	return Artifact{
		FormatVersion: "eml-formal-v1",
		RootID:        rootID,
		Nodes:         nodes,
		Expression:    inspection.Normalized.String(),
		Provenance: &Provenance{
			Source: "concept",
			Concept: &ConceptProvenance{
				Name:            inspection.Name,
				Definition:      inspection.Definition,
				DependencyPaths: clonePaths(inspection.DependencyPaths),
			},
		},
	}, nil
}

// JSON renders an artifact in stable pretty-printed JSON for tooling and later
// downstream consumers.
func (a Artifact) JSON() (string, error) {
	payload, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal formal artifact: %w", err)
	}
	return string(payload), nil
}

func exportNodes(expr ast.Expr) (int, []Node) {
	nextID := 1
	rootID, nodes := collectNodes(expr, &nextID)
	return rootID, nodes
}

func collectNodes(expr ast.Expr, nextID *int) (int, []Node) {
	switch n := expr.(type) {
	case ast.One:
		id := *nextID
		*nextID = *nextID + 1
		return id, []Node{{ID: id, Kind: "one"}}
	case ast.Variable:
		id := *nextID
		*nextID = *nextID + 1
		return id, []Node{{ID: id, Kind: "var", Name: n.Name}}
	case ast.Apply:
		leftID, leftNodes := collectNodes(n.Left, nextID)
		rightID, rightNodes := collectNodes(n.Right, nextID)
		id := *nextID
		*nextID = *nextID + 1
		nodes := make([]Node, 0, len(leftNodes)+len(rightNodes)+1)
		nodes = append(nodes, leftNodes...)
		nodes = append(nodes, rightNodes...)
		nodes = append(nodes, Node{
			ID:      id,
			Kind:    "eml",
			LeftID:  leftID,
			RightID: rightID,
		})
		return id, nodes
	default:
		return 0, nil
	}
}

func clone(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{Span: n.Span}
	case ast.Variable:
		return ast.Variable{Name: n.Name, Span: n.Span}
	case ast.Apply:
		return ast.Apply{
			Left:  clone(n.Left),
			Right: clone(n.Right),
			Span:  n.Span,
		}
	default:
		return nil
	}
}

func clonePaths(paths [][]string) [][]string {
	if len(paths) == 0 {
		return nil
	}
	out := make([][]string, 0, len(paths))
	for _, path := range paths {
		cloned := make([]string, len(path))
		copy(cloned, path)
		out = append(out, cloned)
	}
	return out
}
