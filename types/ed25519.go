package types

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

type PublicKey [ed25519.PublicKeySize]byte
type PrivateKey [ed25519.PrivateKeySize]byte
type Signature [ed25519.SignatureSize]byte

func (a PrivateKey) Public() PublicKey {
	var public PublicKey
	ed := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(ed, a[:])
	copy(public[:], ed.Public().(ed25519.PublicKey))
	return public
}

var FullMask = PublicKey{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

func (a PublicKey) IsEmpty() bool {
	empty := PublicKey{}
	return a == empty
}

func (a PublicKey) EqualMaskTo(b, m PublicKey) bool {
	for i := range a {
		if (a[i] & m[i]) != (b[i] & m[i]) {
			return false
		}
	}
	return true
}

func (a PublicKey) CompareTo(b PublicKey) int {
	return bytes.Compare(a[:], b[:])
}

func (a PublicKey) String() string {
	return fmt.Sprintf("%v", hex.EncodeToString(a[:]))
}

func (a PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

func (a PublicKey) Network() string {
	return "ed25519"
}
