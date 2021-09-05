package mania

import (
	"math"

	"github.com/hndada/gosu/common"
)

const (
	MaxChordPenalty = -0.3 // must be same or greater than -0.5
	MaxTrillBonus   = 0.08

	MaxJackBonus  = 0 // 0.2
	Max2JackBonus = 0 //0.05
	Max2DeltaJack = 120
	maxDeltaJack  = 180

	MinHoldTailStrain  = 0   // 0.05
	MaxHoldTailStrain  = 0.3 // 0.2
	ZeroHoldTailStrain = 0   // 0.1
)

var (
	maxDeltaChord   int64
	maxDeltaTrill   int64
	curveTrillChord common.Segments
	curveJack       common.Segments
	curveTail       common.Segments
)

func init() {
	curveTrillChord = common.NewSegments(
		[]float64{
			0,
			float64(Good.Window + 30),
			float64(Miss.Window + 30)},
		[]float64{
			MaxChordPenalty,
			MaxTrillBonus,
			0})

	curveJack = common.NewSegments(
		[]float64{
			0,
			Max2DeltaJack,
			maxDeltaJack},
		[]float64{
			MaxJackBonus,
			Max2JackBonus,
			0})

	curveTail = common.NewSegments(
		[]float64{
			0,
			float64(Kool.Window),
			float64(Bad.Window) + 50},
		[]float64{
			ZeroHoldTailStrain,
			MinHoldTailStrain,
			MaxHoldTailStrain})

	xValues := curveTrillChord.SolveX(0)
	if len(xValues) != 2 {
		panic("incorrect numbers of xValues")
	}
	maxDeltaChord = int64(math.Round(xValues[0]))
	maxDeltaTrill = int64(math.Round(xValues[1]))

	// maxDeltaJack = int(math.Round(tools.SolveX(beatmap.Curves["Jack"], 0)[0]))
}
