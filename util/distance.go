package util

import (
	"crypto/ed25519"

	"github.com/matrix-org/xproto/types"
)

func LessThan(first, second types.PublicKey) bool {
	for i := 0; i < ed25519.PublicKeySize; i++ {
		if first[i] < second[i] {
			return true
		}
		if first[i] > second[i] {
			return false
		}
	}
	return false
}

// DHTOrdered returns true if the order of A, B and C is
// correct, where A < B < C without wrapping.
func DHTOrdered(a, b, c types.PublicKey) bool {
	return LessThan(a, b) && LessThan(b, c)
}

// DHTWrappedOrdered returns true if the ordering of A, B
// and C is correct, where we may wrap around from C to A.
// This gives us the property of the successor always being
// a+1 and the predecessor being a+sizeofkeyspace.
func DHTWrappedOrdered(a, b, c types.PublicKey) bool {
	ab, bc, ca := LessThan(a, b), LessThan(b, c), LessThan(c, a)
	switch {
	case ab && bc:
		return true
	case bc && ca:
		return true
	case ca && ab:
		return true
	}
	return false
}

func ReverseOrdering(target types.PublicKey, input []types.PublicKey) func(i, j int) bool {
	return func(i, j int) bool {
		return DHTWrappedOrdered(input[i], target, input[j])
	}
}

func ForwardOrdering(target types.PublicKey, input []types.PublicKey) func(i, j int) bool {
	return func(i, j int) bool {
		return DHTWrappedOrdered(target, input[i], input[j])
	}
}
