package depthChecker

import (
	"github.com/gen2brain/beeep"
	"testing"
)

func TestSmth(t *testing.T) {
	err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	if err != nil {
		panic(err)
	}
}
