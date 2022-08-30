package gosu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hndada/gosu/ctrl"
)

func NewVolumeHandler(vol *float64) ctrl.F64Handler {
	play := func() { Sounds.Play("default-hover") }
	return ctrl.F64Handler{
		Handler: ctrl.Handler{
			Keys:       []ebiten.Key{ebiten.Key2, ebiten.Key1},
			PlaySounds: []func(){play, play},
			HoldKey:    -1,
		},
		Min:    0,
		Max:    1,
		Unit:   0.05,
		Target: vol,
	}
}

func NewSpeedHandler(speedBase *float64) ctrl.F64Handler {
	play := func() { Sounds.Play("default-hover") }
	return ctrl.F64Handler{
		Handler: ctrl.Handler{
			Keys:       []ebiten.Key{ebiten.Key4, ebiten.Key3},
			PlaySounds: []func(){play, play},
			HoldKey:    -1,
		},
		Min:    0.1,
		Max:    2,
		Unit:   0.1,
		Target: speedBase,
	}
}

func NewModeTypeHandler(cursor *int, len int) ctrl.IntHandler {
	play := func() { Sounds.Play("default-hover") }
	return ctrl.IntHandler{
		Handler: ctrl.Handler{
			Keys:       []ebiten.Key{ebiten.Key5},
			PlaySounds: []func(){play},
			HoldKey:    -1,
		},
		Min:    0,
		Max:    len,
		Unit:   1,
		Target: cursor,
		Loop:   true,
	}
}

// Todo: should Max be *int?
func NewSelectHandler(cursor *int, len int) ctrl.IntHandler {
	play := func() { Sounds.Play("default-hover") }
	return ctrl.IntHandler{
		Handler: ctrl.Handler{
			Keys:       []ebiten.Key{ebiten.KeyDown, ebiten.KeyUp},
			PlaySounds: []func(){play, play},
			HoldKey:    -1,
		},
		Min:    0,
		Max:    len,
		Unit:   1,
		Target: cursor,
		Loop:   true,
	}
}