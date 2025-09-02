package petrinet

import "slices"

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
func CartestianProduct[T any](input [][]T) [][]T {
	if len(input) == 0 {
		return [][]T{}
	}

	ret := [][]T{{}}

	for _, opts := range input {
		if len(opts) == 0 {
			return [][]T{}
		}

		next := [][]T{}
		for _, r := range ret {
			for _, o := range opts {
				next = append(next, append(slices.Clone(r), o))
			}
		}
		ret = next
	}

	return ret
}
