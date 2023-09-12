package util

import (
	"sort"

	"github.com/matrix-org/xproto/types"
)

type Overlay struct {
	target types.PublicKey
	ourkey types.PublicKey
	keys   []types.PublicKey
}

func (o *Overlay) candidates() ([]types.PublicKey, error) {
	sort.SliceStable(o.keys, ForwardOrdering(o.ourkey, o.keys))

	mustWrap := o.target.CompareTo(o.ourkey) < 0
	hasWrapped := !mustWrap

	cap, last := len(o.keys), o.keys[0]
	for i, k := range o.keys {
		if hasWrapped {
			if k.CompareTo(o.target) > 0 {
				cap = i
				break
			}
		} else {
			hasWrapped = k.CompareTo(last) < 0
		}
	}
	o.keys = o.keys[:cap]

	nc := len(o.keys)
	if nc > 3 {
		nc = 3
	}

	candidates := []types.PublicKey{}
	candidates = append(candidates, o.keys[:nc]...)

	kc := len(o.keys)
	if kc > 10 {
		candidates = append(candidates, o.keys[kc/8])
	}
	if kc > 7 {
		candidates = append(candidates, o.keys[kc/4])
	}
	if kc > 4 {
		candidates = append(candidates, o.keys[kc/2])
	}

	return candidates, nil
}
