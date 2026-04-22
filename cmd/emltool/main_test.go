package main

import (
	"bytes"
	"io"
	"os"
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
