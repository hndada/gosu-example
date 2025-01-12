package plays

import (
	"fmt"
	"math"
	"sort"

	"github.com/hndada/gosu/format/osu"
)

// int32 is enough for dealing with scene time in millisecond.
// Maximum duration with int32 is around 24 days.

// Except Volume, all fields in Dynamic are related in beat, or 'Pace'.
// Tempo: Allergo, Adagio
// Rhythm: confusing with pattern
// Measure: aka BPM
// Meter: confusing with field 'Meter'
type Dynamic struct {
	Time    int32
	BPM     float64
	Speed   float64
	Meter   int
	NewBeat bool // NewBeat draws a bar.

	Volume    float64 // Used when sample volume is 0.
	Highlight bool

	// Position is for drawing notes and bars efficiently. In piano play mode,
	// Only cursor is updated in every Update(), then notes and bars are drawn
	// based on the difference between their positions and cursor's.
	Position float64
}

func (d Dynamic) BeatDuration() float64 {
	return float64(d.Meter) * (60000 / d.BPM)
}

// No two Dynamics have same Time.
type Dynamics struct {
	data       []Dynamic
	idx        int
	span       int32 // total duration of the chart
	SpeedScale float64
	// Reach      float64
}

func NewDynamics(chart any) (Dynamics, error) {
	var ds []Dynamic
	var span int32
	switch chart := chart.(type) {
	case *osu.Format:
		ds = newDynamicListFromOsu(chart)
		span = int32(chart.Duration())
	}

	if len(ds) == 0 {
		return Dynamics{}, fmt.Errorf("no Dynamics in the chart")
	}
	dys := Dynamics{data: ds, span: span}
	dys.setPositions()
	dys.SpeedScale = 1
	return dys, nil
}

// When gathering Dynamics from osu.Format, it should input the whole slice.
// It is because osu.Format.TimingPoints brings some value from previous TimingPoint.
// First BPM is used as temporary main BPM.
func newDynamicListFromOsu(f *osu.Format) []Dynamic {
	sort.SliceStable(f.TimingPoints, func(i int, j int) bool {
		if f.TimingPoints[i].Time == f.TimingPoints[j].Time {
			return f.TimingPoints[i].Uninherited
		}
		return f.TimingPoints[i].Time < f.TimingPoints[j].Time
	})
	// Inherited points without Uninherited points will go dropped.
	for len(f.TimingPoints) > 0 && !f.TimingPoints[0].Uninherited {
		f.TimingPoints = f.TimingPoints[1:]
	}
	if len(f.TimingPoints) == 0 {
		return nil
	}

	tempMainBPM := f.TimingPoints[0].BPM()
	ds := make([]Dynamic, 0, len(f.TimingPoints))
	prevBPM := tempMainBPM
	for _, timingPoint := range f.TimingPoints {
		d := Dynamic{
			Time:      int32(timingPoint.Time),
			BPM:       prevBPM,
			Speed:     prevBPM / tempMainBPM,
			Meter:     timingPoint.Meter,
			NewBeat:   timingPoint.Uninherited,
			Volume:    float64(timingPoint.Volume) / 100,
			Highlight: timingPoint.IsKiai(),
		}
		if timingPoint.Uninherited {
			d.BPM = timingPoint.BPM()
			d.Speed = d.BPM / tempMainBPM
		} else {
			d.Speed *= timingPoint.BeatLengthScale()
		}

		// Drop a Dynamic with a same time.
		if len(ds) > 0 && ds[len(ds)-1].Time == d.Time {
			// Either one makes Dynamic a NewBeat.
			d.NewBeat = ds[len(ds)-1].NewBeat || d.NewBeat
			ds = ds[:len(ds)-1]
		}
		ds = append(ds, d)
		prevBPM = d.BPM
	}
	return ds
}

func (dys Dynamics) Dynamics() []Dynamic { return dys.data }

// BPM with longest duration will be main BPM.
// When there are multiple BPMs with same duration, larger one will be main BPM.
func (dys Dynamics) BPMs() (main, min, max float64) {
	bpmDurations := make(map[float64]int32)
	for i, d := range dys.data {
		if i == 0 {
			bpmDurations[d.BPM] += d.Time
		}
		if i < len(dys.data)-1 {
			bpmDurations[d.BPM] += dys.data[i+1].Time - d.Time
		} else {
			bpmDurations[d.BPM] += dys.span - d.Time // Bounds to final note time; confirmed with test.
		}
	}
	var maxDuration int32
	min = math.MaxFloat64
	for bpm, duration := range bpmDurations {
		if maxDuration < duration {
			maxDuration = duration
			main = bpm
		} else if maxDuration == duration && main < bpm {
			main = bpm
		}
		if min > bpm {
			min = bpm
		}
		if max < bpm {
			max = bpm
		}
	}
	return
}

func (dys *Dynamics) setPositions() {
	// Brilliant idea: Make SpeedScale scaled by MainBPM.
	mainBPM, _, _ := dys.BPMs()
	bpmScale := dys.data[0].BPM / mainBPM
	for i, d := range dys.data {
		d.Speed *= bpmScale
		if i == 0 {
			dys.data[i].Position = float64(d.Time) * d.Speed
			continue
		}
		prev := dys.data[i-1]
		gain := prev.Speed * float64(d.Time-prev.Time)
		dys.data[i].Position = prev.Position + gain
	}
}

func (dys Dynamics) BeatTimes() (times []int32) {
	// These variables are for iterating over the Time.
	var start, end, step float64
	const bufferTime = 5000

	// times before first Dynamic
	start = float64(dys.data[0].Time)
	end = start
	if end > -bufferTime {
		end = -bufferTime
	}
	step = dys.data[0].BeatDuration()
	for t := start; t >= end; t -= step {
		times = append([]int32{int32(t)}, times...)
	}
	// Need to drop a last element because it will be duplicated.
	times = times[:len(times)-1]

	// times after first Dynamic
	var newDys []Dynamic
	for _, d := range dys.data {
		if d.NewBeat {
			newDys = append(newDys, d)
		}
	}

	for i, nd := range newDys {
		start = float64(nd.Time)
		if i == len(newDys)-1 {
			end = float64(dys.span + bufferTime)
		} else {
			end = float64(newDys[i+1].Time)
		}
		step = nd.BeatDuration()
		for t := start; t < end; t += step {
			times = append(times, int32(t))
		}
	}
	return
}

// UpdateIndex update index of Dynamics and returns current Dynamic.
func (dys *Dynamics) UpdateIndex(now int32) Dynamic {
	for i := dys.idx; i < len(dys.data); i++ {
		// if-condition first, then update index.
		if dys.data[i].Time > now {
			break
		}
		dys.idx = i
	}
	return dys.Current()
}

// Closure
// For one-time iteration over data.
func (dys Dynamics) FuncCurrentDynamic() func(now int32) Dynamic {
	i := 0
	return func(now int32) Dynamic {
		// if-condition first, then update index
		// so that index is not updated when Time is larger than now.
		for i := dys.idx; i < len(dys.data); i++ {
			if dys.data[i].Time > now {
				break
			}
			i++
		}
		return dys.data[i]
	}
}

func (dys *Dynamics) Reset() { dys.idx = 0 }

func (dys Dynamics) Current() Dynamic { return dys.data[dys.idx] }

// The unit of speed is logical pixel per millisecond.
func (dys Dynamics) Speed() float64 {
	return dys.Current().Speed * dys.SpeedScale
}

func (dys Dynamics) Position(t int32) float64 {
	dy := dys.Current()
	return dy.Position + dys.Speed()*float64(t-dy.Time)
}

// NoteExposureDuration returns duration of note exposure:
// the time that a note is visible on the screen.

// reach stands for the distance that a note travels.
func (dys Dynamics) NoteExposureDuration(reach float64) int32 {
	speed := dys.Speed()
	if speed == 0 {
		// This is not likely to happen.
		return 1<<31 - 1
	}
	return int32(reach/speed) + 1
}
