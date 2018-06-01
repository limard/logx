package logx

import (
	"testing"
)

func TestTrace(t *testing.T) {
	Trace()
	Debug("Debug")

	l := New("", "Logx.Obj")
	l.Trace()
	l.Debug("Debug")
}