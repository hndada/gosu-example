package ctrl

import (
	"github.com/hndada/gosu/audios"
	"github.com/hndada/gosu/input"
)

var KeysLeftRight = [2]input.Key{input.KeyArrowLeft, input.KeyArrowRight}
var KeysUpDown = [2]input.Key{input.KeyArrowUp, input.KeyArrowDown}

// short: 0.1s
// long: 0.5s

// Todo: Modifiers work strangely when there are plural modifiers.
// Todo: support modifiers for KeyHandler
type KeyHandler struct {
	Handler
	Modifier input.Key // Handler works only when Modifier is pressed.
	Keys     [2]input.Key
	Sounds   [2]audios.SoundPlayer
	Volume   *float64

	holdIndex int
	active    bool
	countdown int // User needs to hold for a while to activate.
}

// Update returns whether the handler has fired (triggered) or not.
func (kh *KeyHandler) Handle() (fired bool) {
	if kh.countdown > 0 {
		kh.countdown--
		return
	}

	if kh.Modifier != input.KeyNone && !input.IsKeyPressed(kh.Modifier) {
		kh.reset()
		return
	}

	if kh.holdIndex > none && !input.IsKeyPressed(kh.Keys[kh.holdIndex]) {
		kh.reset()
	}

	for i, k := range kh.Keys {
		if input.IsKeyPressed(k) {
			kh.holdIndex = i
			break
		}
	}
	if kh.holdIndex == none {
		return
	}

	[]func(){kh.Decrease, kh.Increase}[kh.holdIndex]()

	kh.Sounds[kh.holdIndex].Play(*kh.Volume)

	if kh.active {
		kh.countdown = shortTicks
	} else {
		kh.countdown = longTicks
	}
	kh.active = true

	return true
}

func (kh *KeyHandler) reset() {
	kh.active = false
	kh.holdIndex = none
}
