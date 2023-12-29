package piano

import (
	"github.com/hndada/gosu/game"
)

type StageOpts struct {
	keyCount int
	Ws       map[int]float64
	w        float64
	H        float64 // center bottom
	X        float64
}

func NewStageOpts(keyCount int) StageOpts {
	opts := StageOpts{
		keyCount: keyCount,
		Ws: map[int]float64{
			1:  240,
			2:  260,
			3:  280,
			4:  300,
			5:  320,
			6:  340,
			7:  360,
			8:  380,
			9:  400,
			10: 420,
		},
		H: 0.90 * game.ScreenH,
		X: 0.50 * game.ScreenW,
	}
	opts.w = opts.Ws[opts.keyCount]
	return opts
}