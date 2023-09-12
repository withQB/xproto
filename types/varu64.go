package types

import "fmt"

type Varu64 uint64

func (n Varu64) MarshalBinary(b []byte) (int, error) {
	if len(b) < n.Length() {
		return 0, fmt.Errorf("input slice too small")
	}
	l := n.Length()
	i := l - 1
	b[i] = byte(n & 0x7f)
	for n >>= 7; n != 0; n >>= 7 {
		i--
		b[i] = byte(n | 0x80)
	}
	return l, nil
}

func (n *Varu64) UnmarshalBinary(buf []byte) (int, error) {
	l := 0
	*n = Varu64(0)
	for _, b := range buf {
		*n <<= 7
		*n |= Varu64(b & 0x7f)
		l++
		if b&0x80 == 0 {
			break
		}
	}
	return l, nil
}

func (n Varu64) Length() int {
	l := 1
	for e := n >> 7; e > 0; e >>= 7 {
		l++
	}
	return l
}

func (n Varu64) MinLength() int {
	return 1
}
