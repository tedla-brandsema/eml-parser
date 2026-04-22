package concepts

// StandardLibrary returns the initial grounded concept registry.
//
// The parser remains atomic; this library is where named mathematical
// constructions begin. Only concepts that are directly grounded in the current
// implementation should be registered here.
func StandardLibrary() *Registry {
	return NewRegistry().
		MustRegister(Definition{
			Name: "one",
			Body: ConstOne(),
		}).
		MustRegister(Definition{
			Name: "e",
			Body: EML(ConstOne(), ConstOne()),
		}).
		MustRegister(Definition{
			Name:   "exp",
			Params: []string{"x"},
			Body:   EML(P("x"), ConstOne()),
		}).
		MustRegister(Definition{
			Name:   "log",
			Params: []string{"x"},
			Body:   EML(ConstOne(), Ref("exp", EML(ConstOne(), P("x")))),
		}).
		MustRegister(Definition{
			Name:   "id",
			Params: []string{"x"},
			Body:   Ref("exp", Ref("log", P("x"))),
		}).
		MustRegister(Definition{
			Name: "zero",
			Body: Ref("log", ConstOne()),
		}).
		MustRegister(Definition{
			Name: "minus_one",
			Body: Ref("neg", ConstOne()),
		}).
		MustRegister(Definition{
			Name: "two",
			Body: Ref("add", ConstOne(), ConstOne()),
		}).
		MustRegister(Definition{
			Name: "half",
			Body: Ref("div", ConstOne(), Ref("two")),
		}).
		MustRegister(Definition{
			Name:   "sub",
			Params: []string{"a", "b"},
			Body:   EML(Ref("log", P("a")), Ref("exp", P("b"))),
		}).
		MustRegister(Definition{
			Name:   "neg",
			Params: []string{"x"},
			Body:   Ref("sub", Ref("sub", Ref("e"), P("x")), Ref("e")),
		}).
		MustRegister(Definition{
			Name:   "add",
			Params: []string{"a", "b"},
			Body:   Ref("sub", P("a"), Ref("neg", P("b"))),
		}).
		MustRegister(Definition{
			Name:   "recip",
			Params: []string{"x"},
			Body:   Ref("exp", Ref("neg", Ref("log", P("x")))),
		}).
		MustRegister(Definition{
			Name:   "mul",
			Params: []string{"a", "b"},
			Body:   Ref("exp", Ref("add", Ref("log", P("a")), Ref("log", P("b")))),
		}).
		MustRegister(Definition{
			Name:   "div",
			Params: []string{"a", "b"},
			Body:   Ref("exp", Ref("sub", Ref("log", P("a")), Ref("log", P("b")))),
		}).
		MustRegister(Definition{
			Name:   "pow",
			Params: []string{"a", "b"},
			Body:   Ref("exp", Ref("mul", P("b"), Ref("log", P("a")))),
		}).
		MustRegister(Definition{
			Name:   "square",
			Params: []string{"x"},
			Body:   Ref("mul", P("x"), P("x")),
		}).
		MustRegister(Definition{
			Name:   "sqrt",
			Params: []string{"x"},
			Body:   Ref("pow", P("x"), Ref("half")),
		}).
		MustRegister(Definition{
			Name:   "sinh",
			Params: []string{"x"},
			Body:   Ref("mul", Ref("half"), Ref("sub", Ref("exp", P("x")), Ref("exp", Ref("neg", P("x"))))),
		}).
		MustRegister(Definition{
			Name:   "cosh",
			Params: []string{"x"},
			Body:   Ref("mul", Ref("half"), Ref("add", Ref("exp", P("x")), Ref("exp", Ref("neg", P("x"))))),
		}).
		MustRegister(Definition{
			Name:   "tanh",
			Params: []string{"x"},
			Body:   Ref("div", Ref("sinh", P("x")), Ref("cosh", P("x"))),
		}).
		MustRegister(Definition{
			Name:   "asinh",
			Params: []string{"x"},
			Body:   Ref("log", Ref("add", P("x"), Ref("sqrt", Ref("add", Ref("square", P("x")), ConstOne())))),
		}).
		MustRegister(Definition{
			Name:   "acosh",
			Params: []string{"x"},
			Body: Ref("log",
				Ref("add",
					P("x"),
					Ref("mul",
						Ref("sqrt", Ref("add", P("x"), ConstOne())),
						Ref("sqrt", Ref("sub", P("x"), ConstOne())),
					),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "atanh",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("half"),
				Ref("log",
					Ref("div",
						Ref("add", ConstOne(), P("x")),
						Ref("sub", ConstOne(), P("x")),
					),
				),
			),
		}).
		MustRegister(Definition{
			Name: "i",
			Body: Ref("sqrt", Ref("minus_one")),
		}).
		MustRegister(Definition{
			Name: "pi",
			Body: Ref("neg", Ref("mul", Ref("i"), Ref("log", Ref("minus_one")))),
		}).
		MustRegister(Definition{
			Name:   "sin",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("half"),
				Ref("mul",
					Ref("neg", Ref("i")),
					Ref("sub",
						Ref("exp", Ref("mul", Ref("i"), P("x"))),
						Ref("exp", Ref("neg", Ref("mul", Ref("i"), P("x")))),
					),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "cos",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("half"),
				Ref("add",
					Ref("exp", Ref("mul", Ref("i"), P("x"))),
					Ref("exp", Ref("neg", Ref("mul", Ref("i"), P("x")))),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "tan",
			Params: []string{"x"},
			Body:   Ref("div", Ref("sin", P("x")), Ref("cos", P("x"))),
		}).
		MustRegister(Definition{
			Name:   "asin",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("neg", Ref("i")),
				Ref("log",
					Ref("add",
						Ref("mul", Ref("i"), P("x")),
						Ref("sqrt", Ref("sub", ConstOne(), Ref("square", P("x")))),
					),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "acos",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("neg", Ref("i")),
				Ref("log",
					Ref("add",
						P("x"),
						Ref("mul", Ref("i"), Ref("sqrt", Ref("sub", ConstOne(), Ref("square", P("x"))))),
					),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "atan",
			Params: []string{"x"},
			Body: Ref("mul",
				Ref("mul", Ref("i"), Ref("half")),
				Ref("sub",
					Ref("log", Ref("sub", ConstOne(), Ref("mul", Ref("i"), P("x")))),
					Ref("log", Ref("add", ConstOne(), Ref("mul", Ref("i"), P("x")))),
				),
			),
		}).
		MustRegister(Definition{
			Name:   "sigmoid",
			Params: []string{"x"},
			Body:   Ref("recip", Ref("add", ConstOne(), Ref("exp", Ref("neg", P("x"))))),
		})
}
