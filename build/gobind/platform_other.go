//go:build !ios && !darwin
// +build !ios,!darwin

package gobind

import "log"

type BindLogger struct{}

func (nsl BindLogger) Write(p []byte) (n int, err error) {
	log.Println(string(p))
	return len(p), nil
}
