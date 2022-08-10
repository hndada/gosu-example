package gosu

import (
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hndada/gosu/parse/osr"
)

// ScenePlay: struct, PlayScene: function
// I don't like ScenePlay.Play is scattered in the ScenePlay code.
// There are 5 places that s.Play used in the code.
type ScenePlay struct {
	Chart     *Chart
	PlayNotes []*PlayNote

	Speed        float64
	Tick         int
	FetchPressed func() []bool
	LastPressed  []bool
	Pressed      []bool
	TransPoint
	StagedNotes []*PlayNote

	Combo          int
	Karma          float64
	KarmaSum       float64
	JudgmentCounts []int

	Play        bool // Whether the scene is for play or not
	MusicFile   io.ReadSeekCloser
	MusicPlayer AudioPlayer
	Skin
	Background            Sprite
	LastJudgment          Judgment
	LastJudgmentCountdown int
}

// Put some time of waiting
var (
	WaitBefore int64 = int64(-1.8 * 1000)
	WaitAfter  int64 = 3 * 1000
)

func NewScenePlay(c *Chart, cpath string, rf *osr.Format, play bool) *ScenePlay {
	s := new(ScenePlay)
	s.Chart = c
	s.PlayNotes, s.StagedNotes = NewPlayNotes(c) // Todo: add Mods to input param

	s.Speed = Speed // From global variable
	bufferTime := WaitBefore
	if rf != nil && rf.BufferTime() < bufferTime {
		bufferTime = rf.BufferTime()
	}
	s.Tick = MsecToTick(bufferTime)
	if rf != nil {
		s.FetchPressed = NewReplayListener(rf, s.Chart.KeyCount, bufferTime)
	} else {
		s.FetchPressed = NewListener(KeySettings[s.Chart.KeyCount])
	}
	s.LastPressed = make([]bool, c.KeyCount)
	s.Pressed = make([]bool, c.KeyCount)
	s.TransPoint = TransPoint{
		s.Chart.SpeedFactors[0],
		s.Chart.Tempos[0],
		s.Chart.Volumes[0],
		s.Chart.Effects[0],
	}
	s.Karma = 1
	s.JudgmentCounts = make([]int, 5)

	s.Play = play
	if !s.Play {
		return s
	}
	s.MusicFile, s.MusicPlayer = NewAudioPlayer(c.MusicPath(cpath))
	s.MusicPlayer.SetVolume(Volume)
	s.Skin = SkinMap[c.KeyCount]
	if img := NewImage(s.Chart.BackgroundPath(cpath)); img != nil {
		s.Background = Sprite{
			I: NewImage(s.Chart.BackgroundPath(cpath)),
		}
		s.Background.SetFullscreen()
	} else {
		s.Background = RandomDefaultBackground
	}
	return s
}

// TPS affects only on Update(), not on Draw().
// Todo: keep playing music when making SceneResult
func (s *ScenePlay) Update(g *Game) {
	var maxJudgmentCountdown int = MsecToTick(2250)
	if s.Play && (ebiten.IsKeyPressed(ebiten.KeyEscape) || s.IsFinished()) {
		s.MusicPlayer.Close()
		s.MusicFile.Close()
		g.Scene = selectScene
		return
	}
	if s.Play && s.Tick == 0 {
		s.MusicPlayer.Play()
	}
	for s.SpeedFactor.Next != nil && s.Time() < s.SpeedFactor.Next.Time {
		s.SpeedFactor = s.SpeedFactor.Next
	}
	for s.Tempo.Next != nil && s.Time() < s.Tempo.Next.Time {
		s.Tempo = s.Tempo.Next
	}
	for s.Volume.Next != nil && s.Time() < s.Volume.Next.Time {
		s.Volume = s.Volume.Next
	}
	for s.Effect.Next != nil && s.Time() < s.Effect.Next.Time {
		s.Effect = s.Effect.Next
	}
	s.LastPressed = s.Pressed
	s.Pressed = s.FetchPressed()
	for k, n := range s.StagedNotes {
		if n == nil {
			continue
		}
		if s.Play && n.Type != Tail && s.KeyAction(k) == Hit {
			n.PlaySE()
		}
		td := n.Time - s.Time() // Time difference; negative values means late hit
		if n.Scored {
			if n.Type != Tail {
				panic("non-tail note has not flushed")
			}
			if td < Miss.Window { // Keep Tail being staged until nearly ends
				s.StagedNotes[n.Key] = n.Next
			}
			continue
		}
		if !s.Play {
			continue
		}
		var worst Judgment
		if j := Verdict(n.Type, s.KeyAction(n.Key), td); j.Window != 0 {
			fmt.Printf("Time difference: %d\n", td)
			s.Score(n, j)
			if worst.Window < j.Window {
				worst = j
			}
		}
		if worst.Window != 0 {
			s.LastJudgment = worst
			s.LastJudgmentCountdown = maxJudgmentCountdown
		}
		if s.LastJudgmentCountdown == 0 {
			s.LastJudgment = Judgment{}
		} else {
			s.LastJudgmentCountdown--
		}
	}
	s.Tick++
}
func (s *ScenePlay) Draw(screen *ebiten.Image) {
	bgop := s.Background.Op()
	bgop.ColorM.ChangeHSV(0, 1, BgDimness)
	screen.DrawImage(s.Background.I, bgop)

	s.FieldSprite.Draw(screen)
	s.HintSprite.Draw(screen)
	s.DrawLongNotes(screen)
	s.DrawNotes(screen)
	if s.Combo > 0 {
		s.DrawCombo(screen)
	}
	if s.LastJudgmentCountdown > 0 { // Draw the same judgment for a while.
		for i, j := range Judgments {
			if j.Window == s.LastJudgment.Window {
				s.JudgmentSprites[i].Draw(screen)
				break
			}
		}
	}
	s.DrawScore(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"CurrentFPS: %.2f\nCurrentTPS: %.2f\nTime: %.3fs\n"+
			"Score: %.0f\nKarma: %.2f\nCombo: %d\n"+
			"judge: %v", ebiten.CurrentFPS(), ebiten.CurrentTPS(), float64(s.Time())/1000,
		s.CurrentScore(), s.Karma, s.Combo,
		s.JudgmentCounts))
}

// Each number image has different size.
func (s *ScenePlay) DrawCombo(screen *ebiten.Image) {
	var wsum int
	vs := make([]int, 0)
	for v := s.Combo; v > 0; v /= 10 {
		vs = append(vs, v%10) // Little endian
		wsum += int(s.ComboSprites[v%10].W + ComboGap)
	}
	wsum -= int(ComboGap)
	x := (screenSizeX + float64(wsum)) / 2
	for _, v := range vs {
		x -= s.ComboSprites[v].W + ComboGap
		sprite := s.ComboSprites[v]
		sprite.X = x
		sprite.Y = ComboPosition - sprite.H/2
		sprite.Draw(screen)
	}
}

func (s *ScenePlay) DrawScore(screen *ebiten.Image) {
	var wsum int
	vs := make([]int, 0)
	for v := int(s.CurrentScore()); v > 0; v /= 10 {
		vs = append(vs, v%10) // Little endian
		wsum += int(s.ComboSprites[v%10].W)
	}
	x := float64(screenSizeX)
	for _, v := range vs {
		x -= s.ScoreSprites[v].W
		sprite := s.ScoreSprites[v]
		sprite.X = x
		sprite.Draw(screen)
	}
}

func (s ScenePlay) Time() int64 { // In milliseconds.
	return int64(float64(s.Tick) / float64(MaxTPS) * 1000)
}
func (s ScenePlay) IsFinished() bool {
	return s.Time() > WaitAfter+s.PlayNotes[len(s.PlayNotes)-1].Time
}
func TickToMsec(tick int) int64 { return int64(1000 * float64(tick) / float64(MaxTPS)) }
func MsecToTick(msec int64) int { return int(float64(msec) * float64(MaxTPS) / 1000) }
