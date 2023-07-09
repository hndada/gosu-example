package piano

import "github.com/hndada/gosu/mode"

type Bar struct {
	Time int64 // Times are in milliseconds.

	Position float64
	Next     *Bar
	Prev     *Bar
}

func NewBars(dys []*mode.Dynamic, duration int64) (bs []*Bar) {
	const useDefaultMeter = 0
	times := mode.BeatTimes(dys, duration, useDefaultMeter)

	bs = make([]*Bar, 0, len(times))
	for _, t := range times {
		b := Bar{Time: t}
		bs = append(bs, &b)
	}

	// linking
	var prev *Bar
	for _, b := range bs {
		b.Prev = prev
		if prev != nil {
			prev.Next = b
		}
		prev = b
	}
	return
}
