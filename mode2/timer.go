package mode

import "time"

// TPS affects only on Update(), not on Draw().
var tps = float64(60) // ebitengine's default TPS

func SetTPS(new float64)        { tps = new }
func ToTick(ms int32) int       { return int(tps * float64(ms) / 1000) }
func ToTime(tick int) int32     { return int32(float64(tick) / tps * 1000) }
func ToSecond(ms int32) float64 { return float64(ms) / 1000 }

type Timer struct {
	startTime   time.Time
	pauseTime   time.Time
	paused      bool
	musicOffset int32
	musicPlayed bool // This really matters.
}

func NewTimer(musicOffset int32, wait time.Duration) Timer {
	return Timer{
		startTime:   time.Now().Add(wait),
		musicOffset: musicOffset,
	}
}

func (t Timer) StartTime() time.Time { return t.startTime }

func (t Timer) Now() int32 {
	var duration time.Duration
	if t.paused {
		duration = t.pauseTime.Sub(t.startTime)
	} else {
		duration = time.Since(t.startTime)
	}
	return int32(duration.Milliseconds())
}

func (t Timer) IsPaused() bool { return t.paused }

// No update t.startTime here.
// Notes would look like they suddenly teleport at the beginning.
func (t *Timer) SetMusicPlayed(now time.Time) {
	// offset := time.Duration(t.musicOffset) * time.Millisecond
	// t.startTime = now.Add(-offset)
	t.musicPlayed = true
}

// TL;DR: If you tend to hit early, set positive offset.
// It leads to delayed music / early start time.
func (t *Timer) SetMusicOffset(new int32) {
	// Once the music starts, there isn't much we can do,
	// since music is hard to seek precisely.
	// Instead, we adjust the start time.

	// Positive offset in music infers music is delayed.
	// Delayed music is same as early start time.
	// Hence, as offset increases, start time decreases.
	if t.musicPlayed {
		old := t.musicOffset
		diff := time.Duration(new-old) * time.Millisecond
		t.startTime = t.startTime.Add(-diff)
		t.musicOffset = new
	}
	// If the music has not played yet, we can adjust the offset
	// and let the music played at given delayed time.
	t.musicOffset = new

	// Changing offset might affect to KeyboardState indexing,
	// but it would be negligible because a player tend to hands off the keys
	// when setting offset. Maybe the fact that osu! doesn't allow to change offset
	// during pausing is related to this.
}

func (t *Timer) Pause() {
	t.pauseTime = time.Now()
	t.paused = true
}

func (t *Timer) Resume() {
	elapsedTime := time.Since(t.pauseTime)
	t.startTime = t.startTime.Add(elapsedTime)
	t.paused = false
}

// func (t *Timer) sync() {
// 	const threshold = 30 * 1000
// 	since := int32(time.Since(t.startTime).Milliseconds())
// 	if e := since - t.Now(); e >= threshold {
// 		fmt.Printf("%dms: adjusting time error (%dms)\n", since, e)
// 		t.Tick += ToTick(e)
// 	}
// }