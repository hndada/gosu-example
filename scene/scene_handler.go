package scene

import (
	"github.com/hndada/gosu/audios"
	"github.com/hndada/gosu/ctrl"
	"github.com/hndada/gosu/input"
)

func (s *BaseScene) setMusicVolumeKeyHandler(cfg *Config, asset *Asset) {
	s.MusicVolumeKeyHandler = ctrl.KeyHandler{
		Handler: ctrl.FloatHandler{
			Value: &cfg.MusicVolume,
			Min:   0,
			Max:   1,
			Unit:  0.05,
		},
		Modifier: input.KeyControlLeft,
		Keys:     ctrl.KeysLeftRight,
		Sounds:   asset.ToggleSounds,
		Volume:   &cfg.SoundVolume,
	}
}
func (s *BaseScene) setSoundVolumeKeyHandler(cfg *Config, asset *Asset) {
	s.SoundVolumeKeyHandler = ctrl.KeyHandler{
		Handler: ctrl.FloatHandler{
			Value: &cfg.SoundVolume,
			Min:   0,
			Max:   1,
			Unit:  0.05,
		},
		Modifier: input.KeyAltLeft,
		Keys:     ctrl.KeysLeftRight,
		Sounds:   asset.ToggleSounds,
		Volume:   &cfg.SoundVolume,
	}
}
func (s *BaseScene) setBackgroundBrightnessKeyHandler(cfg *Config, asset *Asset) {
	s.BackgroundBrightnessKeyHandler = ctrl.KeyHandler{
		Handler: ctrl.FloatHandler{
			Value: &cfg.BackgroundBrightness,
			Min:   0,
			Max:   1,
			Unit:  0.1,
		},
		Modifier: input.KeyControlLeft,
		Keys:     [2]input.Key{input.KeyO, input.KeyP},
		Sounds:   asset.ToggleSounds,
		Volume:   &cfg.SoundVolume,
	}
}
func (s *BaseScene) setOffsetKeyHandler(cfg *Config, asset *Asset) {
	s.OffsetKeyHandler = ctrl.KeyHandler{
		Handler: ctrl.Int32Handler{
			Value: &cfg.Offset,
			Min:   -200,
			Max:   200,
			Loop:  false,
			Unit:  1,
		},
		Modifier: input.KeyShiftLeft,
		Keys:     ctrl.KeysLeftRight,
		Sounds:   asset.TransitionSounds,
		Volume:   &cfg.SoundVolume,
	}
}
func (s *BaseScene) setDebugPrintKeyHandler(cfg *Config, asset *Asset) {
	s.DebugPrintKeyHandler = ctrl.KeyHandler{
		Handler: ctrl.BoolHandler{
			Value: &cfg.DebugPrint,
		},
		Modifier: input.KeyNone,
		Keys:     [2]input.Key{input.KeyF12, input.KeyF12},
		Sounds:   [2]audios.SoundPlayer{asset.TapSoundPod, asset.TapSoundPod},
		Volume:   &cfg.SoundVolume,
	}
}
func (s *BaseScene) setSpeedScaleKeyHandlers(cfg *Config, asset *Asset) {
	speedScales := []*float64{&cfg.PianoConfig.SpeedScale}
	s.SpeedScaleKeyHandlers = make([]ctrl.KeyHandler, len(speedScales))
	for i, speedScalePtr := range speedScales {
		s.SpeedScaleKeyHandlers[i] = ctrl.KeyHandler{
			Handler: ctrl.FloatHandler{
				Value: speedScalePtr,
				Min:   0.5,
				Max:   2.5,
				Unit:  0.05,
			},
			Modifier: input.KeyNone,
			Keys:     [2]input.Key{input.KeyPageDown, input.KeyPageUp},
			Sounds:   asset.ToggleSounds,
			Volume:   &cfg.SoundVolume,
		}
	}
}