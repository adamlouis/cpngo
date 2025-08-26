package petrinet_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/stretchr/testify/require"
)

type ExprExample struct {
	Name  string
	Score float64
	Len   int
}

type ExprEnv struct {
	Examples []ExprExample
	Color    any
}

func TestExprLib(t *testing.T) {
	cases := []string{
		`all(Examples, {.Len >= 50})`,
		`len(Examples)`,
		`Examples[0].Name`,
		`Examples[0].Name + "-test-" + Examples[1].Name`,
		`{a: 1, b: "hello world!"}["b"]`,
		`Color`,
		`Color["hi"] == "foobar"`,
		`randb("0.0")`,
		`randb("1.0")`,
		`randb("0.5")`,
		`randb("0.3") `,
		`randf() > 0.5 ? "SUCCESS" : "FAILURE"`,
	}

	randb := expr.Function("randb", func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("expected 1 param, got %d", len(params))
		}
		f, err := strconv.ParseFloat(params[0].(string), 64)
		if err != nil {
			return nil, err
		}
		return f < rand.Float64(), nil
	})
	randf := expr.Function("randf", func(params ...any) (any, error) {
		return rand.Float64(), nil
	})

	for _, code := range cases {
		program, err := expr.Compile(code, randb, randf)
		if err != nil {
			panic(err)
		}

		env := ExprEnv{
			Examples: []ExprExample{{"a", 0.5, 117}, {"b", 0.1, 98}, {"c", 0.7, 112}},
			Color:    map[string]interface{}{"hi": "foobar"},
		}

		output, err := expr.Run(program, env)
		require.NoError(t, err)
		require.NotNil(t, output)
	}
}
