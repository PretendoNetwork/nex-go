package nex

import "golang.org/x/exp/constraints"

func sum[T, O constraints.Integer](data []T) O {
	var result O
	for _, b := range data {
		result += O(b)
	}
	return result
}
