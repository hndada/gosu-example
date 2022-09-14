package drum

// Time is a point of time, duartion a length of time.
// No consider situations that multiple rolls are overlapped.
type Dot struct {
	Floater
	// Showtime int64 // Dot will appear at Showtime.
	RevealTime int64
	First      bool // Whether the dot is the first dot of a Roll note.
	Marked     bool
	Next       *Dot
	Prev       *Dot
}

// Unit of speed is osupixel / 100ms.
// n.SetDots(tp.Speed*speedFactor, bpm)
func NewDots(notes []*Note) (ds []*Dot) {
	for _, n := range notes {
		if n.Type != Roll {
			continue
		}
		var step float64
		if n.Tick >= 2 {
			step = float64(n.Duration) / float64(n.Tick-1)
		}
		for tick := 0; tick < n.Tick; tick++ {
			time := step * float64(tick)
			d := Dot{
				Floater: Floater{
					Time:  n.Time + int64(time),
					Speed: n.Speed,
				},
				First:      tick == 0,
				RevealTime: n.Time - RevealDuration,
			}
			ds = append(ds, &d)
		}
	}

	var prev *Dot
	for _, d := range ds {
		d.Prev = prev
		if prev != nil {
			prev.Next = d
		}
		prev = d
	}
	return
}