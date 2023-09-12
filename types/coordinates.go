package types

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type SwitchPortID Varu64
type Coordinates []SwitchPortID

func (s Coordinates) Network() string {
	return "tree"
}

func (s Coordinates) Len() int {
	return len(s)
}

func (s Coordinates) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Coordinates) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s Coordinates) String() string {
	ports := make([]string, 0, len(s))
	for _, p := range s {
		ports = append(ports, fmt.Sprintf("%d", p))
	}
	return "[" + strings.Join(ports, " ") + "]"
}

func (p Coordinates) MarshalBinary(buf []byte) (int, error) {
	l := 2
	for _, a := range p {
		n, err := Varu64(a).MarshalBinary(buf[l:])
		if err != nil {
			return 0, fmt.Errorf("Varu64(a).MarshalBinary: %w", err)
		}
		l += n
	}
	binary.BigEndian.PutUint16(buf[:2], uint16(l-2))
	return l, nil
}

func (p *Coordinates) UnmarshalBinary(b []byte) (int, error) {
	l := int(binary.BigEndian.Uint16(b[:2]))
	if l == 0 {
		return 2, nil
	}
	if rl := len(b); rl < 2+l {
		return 0, fmt.Errorf("expecting %d bytes but got %d bytes", 2+l, rl)
	}
	ports := make(Coordinates, 0, l)
	read := 2
	b = b[read : l+2]
	for {
		if len(b) < 1 {
			break
		}
		var id Varu64
		l, err := id.UnmarshalBinary(b)
		if err != nil {
			return 0, fmt.Errorf("id.UnmarshalBinary: %w", err)
		}
		ports = append(ports, SwitchPortID(id))
		b = b[l:]
		read += l
	}
	*p = ports
	return read, nil
}

func (p Coordinates) MarshalJSON() ([]byte, error) {
	s := make([]string, 0, len(p))
	for _, id := range p {
		s = append(s, fmt.Sprintf("%d", id))
	}
	return []byte(`"[` + strings.Join(s, " ") + `]"`), nil
}

func (p Coordinates) EqualTo(o Coordinates) bool {
	if len(p) != len(o) {
		return false
	}
	for i := range p {
		if p[i] != o[i] {
			return false
		}
	}
	return true
}

func (a *Coordinates) Copy() Coordinates {
	return append(Coordinates{}, *a...)
}

func (a Coordinates) DistanceTo(b Coordinates) int {
	ancestor := getCommonPrefix(a, b)
	return len(a) + len(b) - 2*ancestor
}

func getCommonPrefix(a, b Coordinates) int {
	c := 0
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			break
		}
		c++
	}
	return c
}
