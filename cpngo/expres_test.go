package cpngo_test

import (
	"fmt"
	"testing"

	"github.com/antonmedv/expr"
)

type ExprExample struct {
	Name  string
	Score float64
	Len   int
}

type ExprEnv struct {
	Examples []ExprExample
}

// TODO(adam): use with arc expressions
func TestExpr(t *testing.T) {
	cases := []string{
		`all(Examples, {.Len >= 50})`,
		`len(Examples)`,
		`Examples[0].Name`,
		`Examples[0].Name + "-test-" + Examples[1].Name`,
	}

	for _, code := range cases {
		program, err := expr.Compile(code, expr.Env(ExprEnv{}))
		if err != nil {
			panic(err)
		}

		env := ExprEnv{
			Examples: []ExprExample{{"a", 0.5, 117}, {"b", 0.1, 98}, {"c", 0.7, 112}},
		}

		output, err := expr.Run(program, env)
		if err != nil {
			panic(err)
		}
		fmt.Println(output)
	}
}
