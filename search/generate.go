package search

import "eml-parser/ast"

// AtomicSeeds returns the smallest raw EML atoms used to seed construction.
func AtomicSeeds(variableNames ...string) []ast.Expr {
	seeds := []ast.Expr{ast.One{}}
	for _, name := range variableNames {
		seeds = append(seeds, ast.Variable{Name: name})
	}
	return seeds
}

// EnumerateBounded generates unique raw EML expressions within the requested
// bounds from the supplied atoms.
func EnumerateBounded(atoms []ast.Expr, bounds Bounds) []ast.Expr {
	if len(atoms) == 0 {
		return nil
	}

	unique := make(map[string]ast.Expr)
	queue := make([]ast.Expr, 0, len(atoms))
	for _, atom := range atoms {
		if !WithinBounds(atom, bounds) {
			continue
		}
		key := CanonicalKey(atom)
		if _, ok := unique[key]; ok {
			continue
		}
		unique[key] = clone(atom)
		queue = append(queue, clone(atom))
	}

	for i := 0; i < len(queue); i++ {
		left := queue[i]
		currentExprs := snapshotValues(unique)
		for _, right := range currentExprs {
			candidates := []ast.Expr{
				ast.Apply{Left: clone(left), Right: clone(right)},
				ast.Apply{Left: clone(right), Right: clone(left)},
			}
			for _, candidate := range candidates {
				if !WithinBounds(candidate, bounds) {
					continue
				}
				key := CanonicalKey(candidate)
				if _, ok := unique[key]; ok {
					continue
				}
				unique[key] = candidate
				queue = append(queue, candidate)
			}
		}
	}

	return snapshotValues(unique)
}

// EnumerateNextLayer generates all unique raw EML expressions of exactly the
// next depth level by pairing each expression in lastLayer with every expression
// in allPrev (which must include lastLayer). The returned expressions all have
// depth = max(depth(lastLayer)) + 1 and are not already present in allPrev.
func EnumerateNextLayer(lastLayer []ast.Expr, allPrev []ast.Expr, bounds Bounds) []ast.Expr {
	seen := make(map[string]bool, len(allPrev))
	for _, e := range allPrev {
		seen[CanonicalKey(e)] = true
	}

	var result []ast.Expr
	for _, a := range lastLayer {
		for _, b := range allPrev {
			for _, expr := range []ast.Expr{
				ast.Apply{Left: clone(a), Right: clone(b)},
				ast.Apply{Left: clone(b), Right: clone(a)},
			} {
				key := CanonicalKey(expr)
				if WithinBounds(expr, bounds) && !seen[key] {
					result = append(result, expr)
					seen[key] = true
				}
			}
		}
	}
	return result
}

func snapshotValues(m map[string]ast.Expr) []ast.Expr {
	out := make([]ast.Expr, 0, len(m))
	for _, expr := range m {
		out = append(out, clone(expr))
	}
	return out
}
