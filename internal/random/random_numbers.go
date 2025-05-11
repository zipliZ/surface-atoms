package random

import (
	"crypto/rand"
	"log/slog"
	"math"
	"math/big"
)

// Float64 generate number from [0, 1). Precision is 12
func Float64() float64 {
	precision := 12
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(math.Pow10(precision))))
	if err != nil {
		slog.Error("Error generating random number", "error", err)
	}

	return float64(randomNumber.Int64()) / math.Pow10(precision)
}
