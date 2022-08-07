package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// ScenePlay: struct, PlayScene: function
type ScenePlay struct {
	Tick int

	Chart       *Chart
	PlayNotes   []*PlayNote
	KeySettings []ebiten.Key // Todo: separate ebiten

	Pressed     []bool
	LastPressed []bool

	StagedNotes    []*PlayNote
	Combo          int
	Karma          float64
	KarmaSum       float64
	JudgmentCounts []int

	// In dev
	ReplayMode   bool
	ReplayStates []ReplayState
	ReplayCursor int

	Speed float64
	TransPoint
	NoteSprites []Sprite
	BodySprites []Sprite
	HeadSprites []Sprite
	TailSprites []Sprite

	MusicFile   io.ReadSeekCloser
	MusicPlayer *audio.Player

	Background      Sprite
	ComboSprites    []Sprite
	ScoreSprites    []Sprite
	JudgmentSprites []Sprite
	ClearSprite     Sprite
	HintSprite      Sprite
}

func TickToMsec(tick int) int64 { return int64(1000 * float64(tick) / float64(MaxTPS)) }
func MsecToTick(msec int64) int { return int(msec) * MaxTPS }
func NewScenePlay(c *Chart, cpath string) *ScenePlay {
	s := new(ScenePlay)
	s.Tick = -2 * MaxTPS // Put 2 seconds of waiting
	s.Chart = c
	s.PlayNotes, s.StagedNotes = NewPlayNotes(c) // Todo: add Mods to input param
	// s.KeySettings = KeySettings[s.Chart.Parameter.KeyCount]
	s.JudgmentCounts = make([]int, 5)
	s.LastPressed = make([]bool, c.KeyCount)
	s.Pressed = make([]bool, c.KeyCount)
	s.Karma = 1
	s.TransPoint = TransPoint{
		s.Chart.SpeedFactors[0],
		s.Chart.Tempos[0],
		s.Chart.Volumes[0],
		s.Chart.Effects[0],
	}
	s.NoteSprites = make([]Sprite, c.KeyCount)
	s.BodySprites = make([]Sprite, c.KeyCount)
	var wsum int
	for k, kind := range NoteKindsMap[c.KeyCount] {
		w := int(NoteWidths[c.KeyCount][int(kind)] * Scale()) // w should be integer, since it is a play note's width.
		var fpath string
		fpath = "skin/note/" + fmt.Sprintf("n%d.png", []int{1, 2, 3, 3}) // Todo: 4th note image
		s.NoteSprites[k] = Sprite{
			I: NewImage(fpath),
			W: float64(w),
			H: NoteHeigth * Scale(),
		}
		fpath = "skin/note/" + fmt.Sprintf("l%d.png", []int{1, 2, 3, 3}) // Todo: 4th note image
		s.BodySprites[k] = Sprite{
			I: NewImage(fpath),
			W: float64(w),
			H: NoteHeigth * Scale(), // Long note body does not have to be scaled though.
		}
		wsum += w
	}
	s.HeadSprites = make([]Sprite, c.KeyCount)
	s.TailSprites = make([]Sprite, c.KeyCount)
	copy(s.HeadSprites, s.NoteSprites)
	copy(s.TailSprites, s.NoteSprites)

	// Todo: Scratch should be excluded to width sum.
	x := (ScreenSizeX - wsum) / 2 // x should be integer as well as w
	for k, kind := range NoteKindsMap[c.KeyCount] {
		s.NoteSprites[k].X = float64(x)
		x += int(NoteWidths[c.KeyCount][kind] * Scale())
	}
	f, err := os.Open(c.MusicPath(cpath))
	if err != nil {
		panic(err)
	}
	s.MusicPlayer, err = Context.NewPlayer(f)
	if err != nil {
		panic(err)
	}

	s.Background = Sprite{
		I: NewImage(c.BgPath(cpath)),
		W: float64(ScreenSizeX),
		H: float64(ScreenSizeY),
	}
	s.ComboSprites = make([]Sprite, 10)
	for i := 0; i < 10; i++ {
		sp := Sprite{
			I: NewImage(fmt.Sprintf("skin/combo/%d.png", i)),
			W: ComboWidth * Scale(),
		}
		sp.H = float64(sp.I.Bounds().Dy()) * (sp.W / float64(sp.I.Bounds().Dx()))
		sp.Y = ComboPosition - sp.H/2
		s.ComboSprites[i] = sp
	}
	s.ScoreSprites = make([]Sprite, 10)
	for i := 0; i < 10; i++ {
		sp := Sprite{
			I: NewImage(fmt.Sprintf("skin/score/%d.png", i)),
			W: ScoreWidth * Scale(),
		}
		sp.H = float64(sp.I.Bounds().Dy()) * (sp.W / float64(sp.I.Bounds().Dx()))
		s.ComboSprites[i] = sp
	}
	s.JudgmentSprites = make([]Sprite, 5)
	for i, name := range []string{"kool", "cool", "good", "bad", "miss"} {
		sp := Sprite{
			I: NewImage(fmt.Sprintf("skin/judgment/%s.png", name)),
			W: JudgmentWidth * Scale(),
		}
		sp.H = float64(sp.I.Bounds().Dy()) * (sp.W / float64(sp.I.Bounds().Dx()))
		sp.X = (float64(ScreenSizeX) - sp.W) / 2
		sp.Y = JudgePosition*Scale() - sp.H/2
		s.JudgmentSprites[i] = sp
	}
	{
		sp := Sprite{
			I: NewImage("skin/play/clear.png"),
		}
		sp.W = float64(sp.I.Bounds().Dx())
		sp.H = float64(sp.I.Bounds().Dy())
		sp.X = (float64(ScreenSizeX) - sp.W) / 2
		sp.Y = (float64(ScreenSizeY) - sp.H) / 2
		s.ClearSprite = sp
	}
	{
		sp := Sprite{
			I: NewImage("skin/play/hint.png"),
		}
		sp.W = float64(wsum)
		sp.H = HintHeight * Scale()
		sp.X = (float64(ScreenSizeX) - sp.W) / 2
		sp.Y = HintPosition*Scale() - sp.H/2
		s.HintSprite = sp
	}
	return s
}

func (s *ScenePlay) Update() {
	if s.IsFinished() {
		if s.MusicPlayer != nil {
			s.MusicFile.Close()
			s.MusicPlayer = nil // Todo: need a test
		}
		return
	}
	s.Tick++
	if s.Tick == 0 {
		s.MusicPlayer.Play()
	}

	for s.Time() < s.SpeedFactor.Next.Time {
		s.SpeedFactor = s.SpeedFactor.Next
	}
	for s.Time() < s.Tempo.Next.Time {
		s.Tempo = s.Tempo.Next
	}
	for s.Time() < s.Volume.Next.Time {
		s.Volume = s.Volume.Next
	}
	for s.Time() < s.Effect.Next.Time {
		s.Effect = s.Effect.Next
	}

	for k, p := range s.Pressed {
		s.LastPressed[k] = p
		if s.ReplayMode {
			for s.ReplayCursor < len(s.ReplayStates)-1 && s.Time() > s.ReplayStates[s.ReplayCursor].Time {
				s.ReplayCursor++
			}
			s.ReplayCursor--
			if s.ReplayCursor < 0 {
				s.ReplayCursor = 0
			}
			s.Pressed = s.ReplayStates[s.ReplayCursor].Pressed
		} else {
			s.Pressed[k] = ebiten.IsKeyPressed(s.KeySettings[k])
		}
	}
	for k, n := range s.StagedNotes {
		if n == nil {
			continue
		}
		if n.Type != Tail && s.KeyAction(k) == Hit {
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
		if j := Verdict(n.Type, s.KeyAction(n.Key), td); j.Window != 0 {
			s.Score(n, j)
		}
	}
}
func (s ScenePlay) Time() int64 {
	return int64(float64(s.Tick) / float64(MaxTPS) * 1000)
}

func (s ScenePlay) IsFinished() bool {
	return s.Time() > 3000+s.PlayNotes[len(s.PlayNotes)-1].Time
}

func (s *ScenePlay) Draw(screen *ebiten.Image) {
	if s.IsFinished() {
		s.DrawClear()
		return
	}
	s.DrawBG()
	s.DrawField()
	s.DrawNotes(screen)
	s.DrawCombo()
	s.DrawJudgment()
	s.DrawOthers() // Score, judgment counts and other states
}

func (s ScenePlay) DrawBG()       {}
func (s ScenePlay) DrawField()    {}
func (s ScenePlay) DrawCombo()    {}
func (s ScenePlay) DrawJudgment() {} // Draw the same judgment for a while.
func (s ScenePlay) DrawScore()    {}
func (s ScenePlay) DrawClear()    {}
func (s ScenePlay) DrawOthers()   {} // judgment counts and scene's state
