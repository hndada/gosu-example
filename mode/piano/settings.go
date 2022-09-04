package piano

import (
	"github.com/hndada/gosu"
	"github.com/hndada/gosu/input"
)

// Logical size of in-game screen
const (
	screenSizeX = gosu.ScreenSizeX
	screenSizeY = gosu.ScreenSizeY
)

var SpeedScale float64 = 1 //.35

var KeySettings = map[int][]input.Key{
	4:               {input.KeyD, input.KeyF, input.KeyJ, input.KeyK},
	5:               {input.KeyD, input.KeyF, input.KeySpace, input.KeyJ, input.KeyK},
	6:               {input.KeyS, input.KeyD, input.KeyF, input.KeyJ, input.KeyK, input.KeyL},
	7:               {input.KeyS, input.KeyD, input.KeyF, input.KeySpace, input.KeyJ, input.KeyK, input.KeyL},
	8 + LeftScratch: {input.KeyA, input.KeyS, input.KeyD, input.KeyF, input.KeySpace, input.KeyJ, input.KeyK, input.KeyL},
	8:               {input.KeyA, input.KeyS, input.KeyD, input.KeyF, input.KeyJ, input.KeyK, input.KeyL, input.KeySemicolon},
	9:               {input.KeyA, input.KeyS, input.KeyD, input.KeyF, input.KeySpace, input.KeyJ, input.KeyK, input.KeyL, input.KeySemicolon},
	10:              {input.KeyA, input.KeyS, input.KeyD, input.KeyF, input.KeyV, input.KeyN, input.KeyJ, input.KeyK, input.KeyL, input.KeySemicolon},
}

// Todo: Should NoteHeight be separated into NoteHeight, HeadHeight, TailHeight?
var (
	ComboPosition    float64 = screenSizeY * 0.40
	JudgmentPosition float64 = screenSizeY * 0.66
	NoteHeigth       float64 = screenSizeY * 0.05 // Applies to all notes
	HitPosition      float64 = screenSizeY * 0.90 // The bottom y-value of Hint,  not a middle or top.
	// HintPosition     float64 = screenSizeY * 0.96
	maxPosition float64 = HitPosition
	minPosition float64 = maxPosition - screenSizeY
	// posMargin should be larger than MaxSize/2 of all note sprites' width or height.
	posMargin   float64 = 100
	bodyLoss    float64 = NoteHeigth // Head/2 + Tail/2.
	ReverseBody bool    = false
)
var NoteWidthsMap = map[int][4]float64{
	4:  {0.065, 0.065, 0.065, 0.065},
	5:  {0.065, 0.065, 0.065, 0.065},
	6:  {0.065, 0.065, 0.065, 0.065},
	7:  {0.06, 0.06, 0.06, 0.06},
	8:  {0.06, 0.06, 0.06, 0.06},
	9:  {0.06, 0.06, 0.06, 0.06},
	10: {0.06, 0.06, 0.06, 0.06},
}
var FieldDark float64 = 0.8 // Todo: FieldDark -> FieldDarkness

const (
	BodySpriteStyleStretch = iota
	BodySpriteStyleAttach
)

// Skin scale settings
// Todo: make the struct SkinScaleSettings
var (
	ComboScale    float64 = 0.75
	ComboDigitGap float64 = screenSizeX * -0.0008
	ScoreScale    float64 = 0.65
	JudgmentScale float64 = 0.33
	HintHeight    float64 = screenSizeY * 0.04

	BodySpriteStyle = BodySpriteStyleStretch
)

// Todo: ratio *= -1 when switching?
func SwitchDirection() {
	max, min := maxPosition, minPosition
	maxPosition = -min
	minPosition = -max
	ReverseBody = !ReverseBody
}
func init() {
	ScaleNoteWidthsMap()
}

// Todo: generalize setting loading function
func ScaleNoteWidthsMap() {
	for k, ws := range NoteWidthsMap {
		ws2 := ws
		for i, w := range ws2 {
			ws2[i] = screenSizeX * w
		}
		NoteWidthsMap[k] = ws2
	}
}

// 1 pixel is 1 millisecond.
func ExposureTime(speed float64) float64 { return (HitPosition + NoteHeigth/2) / speed }
