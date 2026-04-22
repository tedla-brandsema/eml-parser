%{
package parser

import (
	"eml-parser/ast"
	"eml-parser/token"
)
%}

%union {
	tok  token.Token
	expr ast.Expr
}

%token <tok> IDENT ONE EML LPAREN RPAREN COMMA ILLEGAL
%type <expr> input expr

%%

input:
	expr
	{
		$$ = $1
		emllex.(*parserDriver).result = $1
	}

expr:
	ONE
	{
		$$ = ast.One{Span: toASTSpan($1.Span)}
	}
	| IDENT
	{
		$$ = ast.Variable{
			Name: $1.Lexeme,
			Span: toASTSpan($1.Span),
		}
	}
	| EML LPAREN expr COMMA expr RPAREN
	{
		$$ = ast.Apply{
			Left:  $3,
			Right: $5,
			Span: ast.Span{
				Start: toASTPosition($1.Span.Start),
				End:   toASTPosition($6.Span.End),
			},
		}
	}

%%
