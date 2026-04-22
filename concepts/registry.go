package concepts

import (
	"errors"
	"fmt"
	"slices"

	"eml-parser/ast"
)

var (
	ErrDuplicateConcept = errors.New("duplicate concept")
	ErrUnknownConcept   = errors.New("unknown concept")
	ErrArityMismatch    = errors.New("concept arity mismatch")
	ErrUnknownParam     = errors.New("unknown concept parameter")
	ErrConceptCycle     = errors.New("concept cycle detected")
)

// Definition is a named, parameterized concept body.
type Definition struct {
	Name   string
	Params []string
	Body   Expr
}

// Registry stores named concept definitions and expands them to raw EML ASTs.
type Registry struct {
	defs  map[string]Definition
	cache expansionCache
}

func NewRegistry() *Registry {
	return &Registry{
		defs:  make(map[string]Definition),
		cache: newExpansionCache(),
	}
}

// Names returns registered concept names in sorted order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.defs))
	for name := range r.defs {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Definition returns a registered concept definition.
func (r *Registry) Definition(name string) (Definition, bool) {
	def, ok := r.defs[name]
	return def, ok
}

func (r *Registry) Register(def Definition) error {
	if def.Name == "" {
		return fmt.Errorf("concept name cannot be empty")
	}
	if def.Body == nil {
		return fmt.Errorf("concept %q has nil body", def.Name)
	}
	if _, exists := r.defs[def.Name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateConcept, def.Name)
	}

	seen := make(map[string]struct{}, len(def.Params))
	for _, param := range def.Params {
		if param == "" {
			return fmt.Errorf("concept %q has empty parameter name", def.Name)
		}
		if _, exists := seen[param]; exists {
			return fmt.Errorf("concept %q has duplicate parameter %q", def.Name, param)
		}
		seen[param] = struct{}{}
	}

	r.defs[def.Name] = def
	r.cache.clear()
	return nil
}

func (r *Registry) MustRegister(def Definition) *Registry {
	if err := r.Register(def); err != nil {
		panic(err)
	}
	return r
}

// Expand resolves a named concept and returns a raw EML AST.
func (r *Registry) Expand(name string, args ...ast.Expr) (ast.Expr, error) {
	return r.expandCall(name, args, nil)
}

// ExpandSymbolic expands a named concept by binding each parameter to a raw EML
// variable with the same name. This is useful for tooling and inspection.
func (r *Registry) ExpandSymbolic(name string) (ast.Expr, error) {
	if cached, ok := r.cache.getSymbolic(name); ok {
		return cached, nil
	}
	def, ok := r.defs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}
	args := make([]ast.Expr, 0, len(def.Params))
	for _, param := range def.Params {
		args = append(args, ast.Variable{Name: param})
	}
	expanded, err := r.expandCall(name, args, nil)
	if err != nil {
		return nil, err
	}
	r.cache.putSymbolic(name, expanded)
	return cloneAST(expanded), nil
}

// DirectDependencies returns concept names referenced directly in the body.
func (r *Registry) DirectDependencies(name string) ([]string, error) {
	def, ok := r.defs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}
	deps := collectDirectDependencies(def.Body)
	return sortedKeys(deps), nil
}

// TransitiveDependencies returns all concepts required to expand the named
// concept, excluding the concept itself.
func (r *Registry) TransitiveDependencies(name string) ([]string, error) {
	if _, ok := r.defs[name]; !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}
	visited := make(map[string]struct{})
	if err := r.collectTransitive(name, visited, nil); err != nil {
		return nil, err
	}
	delete(visited, name)
	return sortedKeys(visited), nil
}

func (r *Registry) expandCall(name string, args []ast.Expr, stack []string) (ast.Expr, error) {
	def, ok := r.defs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownConcept, name)
	}
	if len(def.Params) != len(args) {
		return nil, fmt.Errorf("%w: %q expects %d args, got %d", ErrArityMismatch, name, len(def.Params), len(args))
	}
	for _, existing := range stack {
		if existing == name {
			return nil, fmt.Errorf("%w: %v -> %s", ErrConceptCycle, stack, name)
		}
	}

	env := make(map[string]ast.Expr, len(def.Params))
	for i, param := range def.Params {
		env[param] = cloneAST(args[i])
	}

	return r.expandExpr(def.Body, env, append(stack, name))
}

func (r *Registry) expandExpr(expr Expr, env map[string]ast.Expr, stack []string) (ast.Expr, error) {
	switch n := expr.(type) {
	case One:
		return ast.One{}, nil
	case Param:
		value, ok := env[n.Name]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownParam, n.Name)
		}
		return cloneAST(value), nil
	case Apply:
		left, err := r.expandExpr(n.Left, env, stack)
		if err != nil {
			return nil, err
		}
		right, err := r.expandExpr(n.Right, env, stack)
		if err != nil {
			return nil, err
		}
		return ast.Apply{Left: left, Right: right}, nil
	case Call:
		args := make([]ast.Expr, 0, len(n.Args))
		for _, arg := range n.Args {
			expanded, err := r.expandExpr(arg, env, stack)
			if err != nil {
				return nil, err
			}
			args = append(args, expanded)
		}
		return r.expandCall(n.Name, args, stack)
	default:
		return nil, fmt.Errorf("unsupported concept expression type %T", expr)
	}
}

func (r *Registry) collectTransitive(name string, visited map[string]struct{}, stack []string) error {
	for _, existing := range stack {
		if existing == name {
			return fmt.Errorf("%w: %v -> %s", ErrConceptCycle, stack, name)
		}
	}
	if _, ok := visited[name]; ok {
		return nil
	}
	visited[name] = struct{}{}

	def := r.defs[name]
	for dep := range collectDirectDependencies(def.Body) {
		if _, ok := r.defs[dep]; !ok {
			return fmt.Errorf("%w: %q", ErrUnknownConcept, dep)
		}
		if err := r.collectTransitive(dep, visited, append(stack, name)); err != nil {
			return err
		}
	}
	return nil
}

func collectDirectDependencies(expr Expr) map[string]struct{} {
	out := make(map[string]struct{})
	walkConceptExpr(expr, func(call Call) {
		out[call.Name] = struct{}{}
	})
	return out
}

func walkConceptExpr(expr Expr, visit func(Call)) {
	switch n := expr.(type) {
	case One, Param:
		return
	case Apply:
		walkConceptExpr(n.Left, visit)
		walkConceptExpr(n.Right, visit)
	case Call:
		visit(n)
		for _, arg := range n.Args {
			walkConceptExpr(arg, visit)
		}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func cloneAST(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{Span: n.Span}
	case ast.Variable:
		return ast.Variable{Name: n.Name, Span: n.Span}
	case ast.Apply:
		return ast.Apply{
			Left:  cloneAST(n.Left),
			Right: cloneAST(n.Right),
			Span:  n.Span,
		}
	default:
		return nil
	}
}

func (r *Registry) symbolicCacheSize() int {
	return len(r.cache.symbolic)
}
