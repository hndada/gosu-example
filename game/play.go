package game

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hndada/gosu/framework/audios"
	"github.com/hndada/gosu/framework/input"
)

func SetTitle(c ChartHeader) {
	title := fmt.Sprintf("gosu | %s - %s [%s] (%s) ", c.Artist, c.MusicName, c.ChartName, c.Charter)
	ebiten.SetWindowTitle(title)
}

// Time is a point of time, duration a length of time.
func TimeToTick(time int64) int { return int(float64(time) / 1000 * float64(TPS)) }
func TickToTime(tick int) int64 { return int64(float64(tick) / float64(TPS) * 1000) }

const Wait = 1800

type Timer struct {
	StartTime time.Time
	Offset    int64
	// Duration  time.Duration
	Tick    int
	MaxTick int // A tick corresponding to EndTime = Duration + WaitAfter
	Now     int64
	Pause   bool
}

func NewTimer(duration int64) Timer {
	return Timer{
		StartTime: time.Now().Add(Wait * time.Millisecond),
		Offset:    int64(Offset),
		// Duration:  time.Duration(duration+2*Wait) * time.Millisecond,
		Tick:    TimeToTick(-Wait),
		MaxTick: TimeToTick(duration + Wait),
		Now:     -Wait,
	}
}

//	func (t Timer) Done() bool {
//		return ebiten.IsKeyPressed(input.KeyEscape) || t.Tick >= t.MaxTick // time.Since(t.StartTime) >= t.Duration
//	}
func (t Timer) Done() bool { return ebiten.IsKeyPressed(input.KeyEscape) }

// func (t *Timer) SwitchPause() {}
func (t *Timer) Ticker() {
	if inpututil.IsKeyJustPressed(input.KeyTab) {
		t.Pause = !t.Pause
	}
	if t.Pause {
		return
	}
	t.Tick++
	// Real-time offset adjusting.
	if td := int64(Offset) - t.Offset; td != 0 {
		t.Offset += td
		t.Tick += TimeToTick(td)
	}
	if t.Now > 0 && ebiten.ActualTPS() < 0.8*float64(TPS) {
		t.Sync()
	}
	t.Now = TickToTime(t.Tick)
}
func (t *Timer) Sync() {
	since := time.Since(t.StartTime).Milliseconds() // - Wait
	if e := since - t.Now; e >= 1 {
		fmt.Printf("adjusting time error at %dms: %d\n", t.Now, e)
		t.Tick += TimeToTick(e)
	}
	// t.Now = TickToTime(t.Tick)
}

// func (t Timer) Time() int64 {
// 	return time.Since(t.StartTime).Milliseconds()
// }

// func (t *Timer) Ticker() {
// 	t.Tick++
// 	since := time.Since(t.StartTime).Milliseconds() + WaitBefore
// 	if e := since - t.Time; e >= 1 {
// 		fmt.Printf("adjusting time error at %dms: %d\n", t.Time, e)
// 		t.Tick += TimeToTick(e)
// 	}
// 	t.Time = TickToTime(t.Tick)
// }

type MusicPlayer struct {
	*Timer
	Volume float64
	Player *audio.Player
	Closer func() error
	pause  bool
}

func NewMusicPlayer(path string, timer *Timer) (MusicPlayer, error) {
	player, closer, err := audios.NewPlayer(path)
	if err != nil {
		return MusicPlayer{}, err
	}
	player.SetVolume(MusicVolume)
	// player.SetBufferSize(100 * time.Millisecond)
	return MusicPlayer{
		Timer:  timer,
		Volume: MusicVolume,
		Player: player,
		Closer: closer,
	}, nil
}

// func (mp MusicPlayer) Play() {
// 	if mp.Player == nil {
// 		return
// 	}
// 	mp.Player.Play()
// }

func (p *MusicPlayer) Update() {
	if p.Player == nil {
		return
	}
	if p.Volume != MusicVolume {
		p.Volume = MusicVolume
		p.Player.SetVolume(p.Volume)
	}
	if p.Timer.Pause {
		if !p.pause {
			p.Player.Pause()
			p.pause = true
		}
	} else {
		if p.pause {
			p.Player.Play()
			p.pause = false
		}
	}
	if p.pause {
		return
	}

	if p.Now == 0+p.Offset {
		p.Player.Play()
	}
	// if p.Now == 150+p.Offset {
	// 	p.Player.Seek(time.Duration(150) * time.Millisecond)
	// }
	// if p.Done() {
	// 	p.Close()
	// }

	// Calling SetVolume in every Update is fine, confirmed by the author, by the way.
	// p.Player.SetVolume(MusicVolume)
}
func (p MusicPlayer) Close() {
	if p.Player != nil {
		p.Player.Close()
		p.Closer()
	}
}

// Todo: need to refactor
// type EffectPlayer struct {
// 	VolumeHandler ctrl.F64Handler
// 	Effects       audios.SoundMap // A player for sample sound is generated at a place.
// }

// func NewEffectPlayer(evh ctrl.F64Handler) EffectPlayer {
// 	return EffectPlayer{
// 		VolumeHandler: evh,
// 		Effects:       audios.NewSoundMap(evh.Target),
// 	}
// }

type KeyLogger struct {
	FetchPressed func() []bool
	LastPressed  []bool
	Pressed      []bool
}

func NewKeyLogger(keySettings []input.Key) (k KeyLogger) {
	keyCount := len(keySettings)
	k.FetchPressed = input.NewListener(keySettings)
	k.LastPressed = make([]bool, keyCount)
	k.Pressed = make([]bool, keyCount)
	return
}
func (l KeyLogger) KeyAction(k int) input.KeyAction {
	return input.CurrentKeyAction(l.LastPressed[k], l.Pressed[k])
}