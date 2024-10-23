package protocol

import (
	"math"
	"testing"
)

func Test_getVolume(t *testing.T) {
	f := float32(1.03)
	n := math.Float32bits(f)
	t.Log(n)

	t.Log(getVolume(n))
}
