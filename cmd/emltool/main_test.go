package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunList(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"list"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	if !strings.Contains(buf.String(), "exp\n") {
		t.Fatalf("expected exp in output, got %q", buf.String())
	}
}

func TestRunExpand(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"expand", "exp"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "eml(x, 1)" {
		t.Fatalf("unexpected expand output: %q", buf.String())
	}
}

func TestRunStats(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"stats", "exp"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"concept: exp",
		"nodes: 3",
		"depth: 2",
		"leaves: 2",
		"direct_dependencies: 0",
		"transitive_dependencies: 0",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunNormalize(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"normalize", "id"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "x" {
		t.Fatalf("unexpected normalize output: %q", buf.String())
	}
}

func TestRunAnalyze(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"analyze", "id"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"concept: id",
		"key: x",
		"nodes: 1",
		"depth: 1",
		"leaves: 1",
		"normalized: x",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunInspect(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"inspect", "id"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"concept: id",
		"definition: id(x) := exp(log(x))",
		"expanded: eml(eml(1, eml(eml(1, x), 1)), 1)",
		"normalized: x",
		"node_delta: 8",
		"depth_delta: 4",
		"dependency_paths:",
		"- id -> exp",
		"- id -> log -> exp",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunSearchReal(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"search-real", "exp_real_small"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"fixture: exp_real_small",
		"diagnostics:",
		"generated:",
		"unique:",
		"duplicates:",
		"normalization_hits:",
		"evaluation_rejects:",
		"scored:",
		"returned:",
		"best_score:",
		"worst_score:",
		"mean_score:",
		"top_candidates:",
		"1. score=",
		"expr=eml(x, 1)",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunGenFamilyArtifacts(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd))
	outDir := filepath.Join(projectRoot, "artifacts", "equivalence")

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}
		if err := os.Remove(filepath.Join(outDir, entry.Name())); err != nil {
			t.Fatalf("Remove failed: %v", err)
		}
	}
	t.Cleanup(func() {
		entries, _ := os.ReadDir(outDir)
		for _, entry := range entries {
			if entry.Name() == ".gitkeep" {
				continue
			}
			_ = os.Remove(filepath.Join(outDir, entry.Name()))
		}
	})

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("os.Chdir failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	if err := run([]string{"gen-family-artifacts"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "generated: 4") {
		t.Fatalf("unexpected command output: %q", out)
	}
	for _, name := range []string{"raw_identity.json", "concept_exp.json", "raw_exp_exp.json", "concept_sigmoid.json"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected artifact %q: %v", name, err)
		}
	}
}

func TestRunGenEquivalenceFamilies(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd))
	outDir := filepath.Join(projectRoot, "artifacts", "equivalence")

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}
		if err := os.Remove(filepath.Join(outDir, entry.Name())); err != nil {
			t.Fatalf("Remove failed: %v", err)
		}
	}
	t.Cleanup(func() {
		entries, _ := os.ReadDir(outDir)
		for _, entry := range entries {
			if entry.Name() == ".gitkeep" {
				continue
			}
			_ = os.Remove(filepath.Join(outDir, entry.Name()))
		}
	})

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	if err := run([]string{"gen-equivalence-families"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "generated: 3") {
		t.Fatalf("unexpected output: %q", out)
	}
	for _, name := range []string{
		"identity_exact.family.json",
		"identity_normalized.family.json",
		"exp_known_equivalence.family.json",
	} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected generated file %q: %v", name, err)
		}
	}
}

func TestRunFormalize(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"formalize", "id"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		`"format_version": "eml-formal-v1"`,
		`"expression": "x"`,
		`"source": "concept"`,
		`"name": "id"`,
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunExperiment(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}
	if err := os.Chdir("/home/ted/projects/go/eml-parser"); err != nil {
		t.Fatalf("os.Chdir failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	specPath := filepath.Join("experiments", "specs", "example_exp_real_grid.json")
	_ = os.Remove(filepath.Join("experiments", "datasets", "example_exp_real_grid.json"))
	_ = os.Remove(filepath.Join("experiments", "results", "example_exp_real_grid.json"))

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"run-experiment", specPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"experiment: example_exp_real_grid",
		"dataset:",
		"result:",
		"recovery_status: exact_normalized_recovery",
		"diagnostics:",
		"top_candidates:",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}

func TestRunReportSuite(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}
	if err := os.Chdir("/home/ted/projects/go/eml-parser"); err != nil {
		t.Fatalf("os.Chdir failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	specPath := filepath.Join("experiments", "specs", "example_exp_real_grid.json")
	resultPath := filepath.Join("experiments", "results", "example_exp_real_grid.json")
	_ = os.Remove(filepath.Join("experiments", "datasets", "example_exp_real_grid.json"))
	_ = os.Remove(resultPath)
	_ = os.Remove(filepath.Join("experiments", "reports", "smoke_suite.json"))
	_ = os.Remove(filepath.Join("experiments", "reports", "smoke_suite.md"))

	if err := run([]string{"run-experiment", specPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if err := run([]string{"report-suite", "smoke_suite", resultPath}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	out := buf.String()
	for _, expected := range []string{
		"suite: smoke_suite",
		"json:",
		"markdown:",
		"total_experiments: 1",
		"success_count: 1",
		"failure_count: 0",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output, got %q", expected, out)
		}
	}
}
