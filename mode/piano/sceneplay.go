package piano

import (
	"io/fs"
	"time"

	"github.com/hndada/gosu/audios"
	"github.com/hndada/gosu/draws"
	"github.com/hndada/gosu/format/osr"
	"github.com/hndada/gosu/input"
	"github.com/hndada/gosu/mode"
)

type ScenePlay struct {
	*Config
	mode.Timer
	now int32 // Use a certain time point. Each Now() may yield different time point.
	input.Keyboard

	// state
	*Chart
	Scorer
	Dynamic *mode.Dynamic

	// audio
	audios.MusicPlayer
	audios.SoundMap

	// draw
	*Asset
	speedScale   float64
	cursor       float64
	highestBar   *Bar
	highestNotes []*Note

	// draw: animation or transition
	keyTimers          []draws.Timer
	noteTimers         []draws.Timer
	keyLightingTimers  []draws.Timer
	hitLightingTimers  []draws.Timer
	holdLightingTimers []draws.Timer
	judgmentTimer      draws.Timer
	comboTimer         draws.Timer

	drawScore func(draws.Image)
	drawCombo func(draws.Image)
}

func NewScenePlay(cfg *Config, assets map[int]*Asset, fsys fs.FS, name string, mods Mods, rf *osr.Format) (s *ScenePlay, err error) {
	s = &ScenePlay{Config: cfg}

	const wait = 1800 * time.Millisecond
	s.Timer = mode.NewTimer(wait)
	s.now = s.Now()

	if rf != nil {
		s.Keyboard = mode.NewReplayPlayer(rf, s.KeyCount)
	} else {
		// keys := input.NamesToKeys(s.KeySettings[s.KeyCount])
		// kb = input.NewKeyboardListener(keys, wait)
	}

	// state
	s.Chart, err = NewChart(cfg, fsys, name, mods)
	if err != nil {
		return
	}
	s.Scorer = NewScorer(s.Chart)
	s.Dynamic = s.Chart.Dynamics[0]

	// audio
	const ratio = 1
	s.MusicPlayer, err = audios.NewMusicPlayerFromFile(fsys, s.MusicFilename, ratio)
	if err != nil {
		return
	}
	s.SoundMap = audios.NewSoundMap(fsys, s.SoundVolume)

	// draw
	s.Asset = assets[s.KeyCount]
	s.speedScale = s.SpeedScale
	s.cursor = float64(s.now) * s.SpeedScale
	s.highestBar = s.Chart.Bars[0]
	s.highestNotes = s.stagedNotes

	// Since timers are now updated in Draw(), their ticks would be dependent on FPS.
	// However, so far TPS and FPS goes synced by SyncWithFPS().
	s.keyTimers = s.newTimers(mode.ToTick(30), 0)
	s.noteTimers = s.newTimers(0, mode.ToTick(400))
	s.keyLightingTimers = s.newTimers(mode.ToTick(30), 0)
	s.hitLightingTimers = s.newTimers(mode.ToTick(150), mode.ToTick(150))
	s.holdLightingTimers = s.newTimers(0, mode.ToTick(300))
	s.judgmentTimer = draws.NewTimer(mode.ToTick(250), mode.ToTick(40))
	s.comboTimer = draws.NewTimer(mode.ToTick(2000), 0)

	const comboBounce = 0.85
	s.drawScore = mode.NewDrawScoreFunc(s.ScoreSprites, &s.Score, s.ScoreSpriteScale)
	s.drawCombo = mode.NewDrawComboFunc(s.ComboSprites, &s.Combo, &s.comboTimer, s.ComboDigitGap, comboBounce)
	return
}

func (s ScenePlay) newTimers(maxTick, period int) []draws.Timer {
	timers := make([]draws.Timer, s.Chart.KeyCount)
	for k := range timers {
		timers[k] = draws.NewTimer(maxTick, period)
	}
	return timers
}

func (s ScenePlay) ChartHeader() mode.ChartHeader { return s.Chart.ChartHeader }
func (s ScenePlay) WindowTitle() string           { return s.Chart.WindowTitle() }
func (s ScenePlay) Speed() float64                { return s.Dynamic.Speed * s.SpeedScale }

// SetMusicVolume()

// Need to re-calculate positions when Speed has changed.
func (s *ScenePlay) SetSpeedScale() {
	c := s.Chart
	old := s.speedScale
	new := s.SpeedScale
	s.cursor *= new / old
	for _, d := range c.Dynamics {
		d.Position *= new / old
	}
	for _, n := range c.Notes {
		n.Position *= new / old
	}
	for _, b := range c.Bars {
		b.Position *= new / old
	}
	s.speedScale = s.SpeedScale
}

func (s *ScenePlay) Update() any {
	s.now = s.Now()
	kas := s.Keyboard.Fetch(s.now)

	// state
	s.Scorer.Update(s.now, kas)
	s.Dynamic = mode.NextDynamics(s.Dynamic, s.now)

	// audio
	if s.now >= 0 && s.now < 300 {
		s.MusicPlayer.Play()
	}
	// Play sounds from one KeyboardAction for simplicity.
	s.playSounds(kas[0])

	// draw
	s.updateCursor()
	s.updateHighestBar()
	s.updateHighestNotes()
	s.ticker()
	return nil
}

// Todo: set all sample volumes in advance?
func (s ScenePlay) playSounds(ka input.KeyboardAction) {
	for k, n := range s.stagedNotes {
		a := ka.Action[k]
		if n.Type != Tail && a == input.Hit {
			name := n.Sample.Filename
			vol := n.Sample.Volume
			if vol == 0 {
				vol = s.Dynamic.Volume
			}
			scale := *s.SoundVolume
			s.SoundMap.Play(name, vol*scale)
		}
	}
}

// When speed changes from fast to slow, which means there are more bars
// on the screen, updateHighestBar() will handle it optimally.
// When speed changes from slow to fast, which means there are fewer bars
// on the screen, updateHighestBar() actually does nothing, which is still
// fine because that makes some unnecessary bars are drawn.
// The same concept also applies to notes.
func (s *ScenePlay) updateHighestBar() {
	upperBound := s.cursor + s.ScreenSize.Y + 100
	for b := s.highestBar; b.Position < upperBound; b = b.Next {
		s.highestBar = b
		if b.Next == nil {
			break
		}
	}
}

func (s *ScenePlay) updateHighestNotes() {
	upperBound := s.cursor + s.ScreenSize.Y + 100
	for k, n := range s.highestNotes {
		for ; n.Position < upperBound; n = n.Next {
			s.highestNotes[k] = n
			if n.Next == nil {
				break
			}
		}
		// Head cannot be the highest note, since drawLongNoteBody
		// is drawn by its Tail.
		if n.Type == Head {
			s.highestNotes[k] = n.Next
		}
	}
}

func (s *ScenePlay) updateCursor() {
	duration := float64(s.now - s.Dynamic.Time)
	s.cursor = s.Dynamic.Position + duration*s.Speed()
}

func (s *ScenePlay) Pause() {
	s.Timer.Pause()
	s.MusicPlayer.Pause()
	s.Keyboard.Pause()
}

func (s *ScenePlay) Resume() {
	s.Timer.Resume()
	s.MusicPlayer.Resume()
	s.Keyboard.Resume()
}

func (s ScenePlay) Finish() any {
	s.MusicPlayer.Close()
	s.Keyboard.Close()
	return s.Scorer
}