package game

import "image"

var Settings struct {
	ScreenSize         image.Point
	JudgmentMeterScale float64 // 1ms 당 2px
	ScoreHeight        float64 // todo: ScoreHeight -> ScoreImageScale
	ComboHeight        float64 // todo: mode마다 combo 위치가 다르므로 game package에 있으면 안됨
	ComboPosition      float64
	ComboGap           float64
}

func init() {
	Settings.JudgmentMeterScale = 2
	Settings.ScoreHeight = 7
	Settings.ComboHeight = 10
	Settings.ComboPosition = 40
	Settings.ComboGap = 0.8
}

func DisplayScale() float64 {
	return float64(Settings.ScreenSize.Y) / 100
}