package entities

import (
	"fmt"
	"math"
	"strconv"
)

func hasAtMostXDecimals(f float64, x int64) bool {
	s := fmt.Sprintf(fmt.Sprintf("%%.%df", x), f)
	converted, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}

	epsilon := 1e-9
	return math.Abs(f-converted) < epsilon
}
