package petrinet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCartesianProduct(t *testing.T) {

	// Input:
	//   - [A1, A2, A3]
	//   - [B1, B2]
	//   - [C1]
	//
	// Output:
	//   - A1, B1, C1
	//   - A1, B2, C1
	//   - A2, B1, C1
	//   - A2, B2, C1
	//   - A3, B1, C1
	//   - A3, B2, C1

	input := [][]string{
		{"A1", "A2", "A3"},
		{"B1", "B2"},
		{"C1"},
	}

	expected := [][]string{
		{"A1", "B1", "C1"},
		{"A1", "B2", "C1"},
		{"A2", "B1", "C1"},
		{"A2", "B2", "C1"},
		{"A3", "B1", "C1"},
		{"A3", "B2", "C1"},
	}

	actual := CartestianProduct(input)

	require.Equal(t, expected, actual)

}
