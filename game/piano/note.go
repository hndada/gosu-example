package piano

import (
	"fmt"
	"image/color"
	"io/fs"
	"sort"

	"github.com/hndada/gosu/draws"
	"github.com/hndada/gosu/format/osu"
	"github.com/hndada/gosu/game"
)

const (
	Normal = iota
	Head
	Tail
	Body
)

type Note struct {
	Time     int32
	Type     int
	Key      int
	Sample   game.Sample
	Duration int32

	// valid    bool    // Indcate that note is not blank.
	position float64 // Scaled x or y value.
	next     int     // For updating staged notes.
	prev     int     // For accessing to Head from Tail.
	scored   bool
}

// The length of the returned slice is 1 or 2.
func newNoteFromOsu(f osu.HitObject, keyCount int) (ns []Note) {
	n := Note{
		Time:   int32(f.Time),
		Type:   Normal,
		Key:    f.Column(keyCount),
		Sample: game.NewSample(f),
		// valid:  true,
	}
	if f.NoteType&osu.ComboMask == osu.HitTypeHoldNote {
		n.Type = Head
		n.Duration = int32(f.EndTime) - n.Time
		n2 := Note{
			Time: n.Time + n.Duration,
			Type: Tail,
			Key:  n.Key,
			// Sample: game.Sample{}, // Tail has no sample sound.
			// Duration: n.Duration, // Todo: 0 or n.Duration?
			// valid: true,
		}
		ns = append(ns, n, n2)
	} else {
		ns = append(ns, n)
	}
	return ns
}

type Notes struct {
	notes        []Note
	keyCount     int
	keysFocus    []int // indexes of focused notes
	sampleBuffer []game.Sample
}

func NewNotes(chart any, dys game.Dynamics) Notes {
	var ns []Note
	var keyCount int
	switch chart := chart.(type) {
	case *osu.Format:
		ns = make([]Note, 0, len(chart.HitObjects)*2)
		for _, ho := range chart.HitObjects {
			ns = append(ns, newNoteFromOsu(ho, keyCount)...)
		}
		keyCount = int(chart.CircleSize)
	}

	sort.Slice(ns, func(i, j int) bool {
		if ns[i].Time == ns[j].Time {
			return ns[i].Key < ns[j].Key
		}
		return ns[i].Time < ns[j].Time
	})

	// Position calculation is based on Dynamics.
	// Farther note has larger position.
	for i, n := range ns {
		d := dys.UpdateIndex(n.Time)
		ns[i].position = d.Position + d.Speed*float64(n.Time-d.Time)

		// Tail's Position should be always equal or larger than Head's.
		if n.Type == Tail {
			head := ns[n.prev]
			if n.position < head.position {
				ns[i].position = head.position
			}
		}
	}

	// linking
	prevs := make([]int, keyCount)
	exists := make([]bool, keyCount)
	for i, n := range ns {
		prev := prevs[n.Key]
		ns[i].prev = prev
		if exists[n.Key] {
			ns[prev].next = i
		}
		prevs[n.Key] = i
		exists[n.Key] = true
	}
	// Set each last note's next with the none value: length of notes.
	for k, prev := range prevs {
		if !exists[k] {
			continue
		}
		ns[prev].next = len(ns)
	}

	// Set key focus indexes.
	// Initialize with the max none value: length of notes.
	kfi := make([]int, keyCount)
	for k := range kfi {
		kfi[k] = len(ns)
	}
	for k := range kfi {
		for i, n := range ns {
			if k == n.Key {
				kfi[n.Key] = i
				break
			}
		}
	}

	return Notes{ns, keyCount, kfi, nil}
}

func (ns Notes) NoteCounts() []int {
	counts := make([]int, 2)
	for _, n := range ns.notes {
		switch n.Type {
		case Normal:
			counts[0]++
		case Head:
			counts[1]++
		}
	}
	return counts
}

func (ns Notes) Span() int32 {
	if len(ns.notes) == 0 {
		return 0
	}
	last := ns.notes[len(ns.notes)-1]
	// No need to add last.Duration, since last is
	// always either Normal or Tail.
	return last.Time
}

type NotesResources struct {
	framesList [4]draws.Frames
}

// When note/normal image is not found, use default's note/normal.
// When note/head image is not found, use user's note/normal.
// When note/tail image is not found, let it be blank.
// When note/body image is not found, use user's note/normal.
// Todo: remove key kind folders
func (res *NotesResources) Load(fsys fs.FS) {
	for nt, ntn := range []string{"normal", "head", "tail", "body"} {
		name := fmt.Sprintf("piano/note/one/%s.png", ntn)
		res.framesList[nt] = draws.NewFramesFromFile(fsys, name)
	}
}

type NotesOptions struct {
	TailOffset int32

	keyCount int
	kw       []float64
	H        float64 // Applies to all types of notes.
	kx       []float64
	y        float64 // center bottom
	keyOrder []KeyKind
	Colors   [4]color.NRGBA
	// LongBodyStyle     int // Stretch or Attach.
}

func NewNotesOptions(stage StageOptions, keys KeysOptions) NotesOptions {
	return NotesOptions{
		TailOffset: 0,

		keyCount: stage.keyCount,
		kw:       keys.kw,
		H:        20,
		kx:       keys.kx,
		y:        keys.y,

		keyOrder: keys.Order(),
		// Colors are from each note's second outermost pixel.
		Colors: [4]color.NRGBA{
			{255, 255, 255, 255}, // One: white
			{239, 191, 226, 255}, // Two: pink
			{218, 215, 103, 255}, // Mid: yellow
			{224, 0, 0, 255},     // Tip: red
		},
		// LongBodyStyle: 0,
	}
}

type NotesComponent struct {
	notes       Notes
	cursor      float64
	keysHolding []bool
	keysLowest  []int // indexes of lowest notes
	keysAnims   [][4]draws.Animation
	h           float64 // used for drawLongNoteBody
	keysColor   []color.NRGBA
}

func NewNotesComponent(res NotesResources, opts NotesOptions, ns Notes, dys game.Dynamics) (cmp NotesComponent) {
	for i, n := range ns.notes {
		if n.Type != Tail {
			continue
		}

		// Apply TailOffset to Tail's Position.
		// Tail's Position should be always equal or larger than Head's.
		d := dys.UpdateIndex(n.Time)
		ns.notes[i].position += float64(opts.TailOffset) * d.Speed
		if head := ns.notes[n.prev]; n.position < head.position {
			ns.notes[i].position = head.position
		}

		// Apply dynamics' volume to note's sample with blank volume.
		if n.Sample.Volume == 0 {
			ns.notes[i].Sample.Volume = d.Volume
		}
	}
	cmp.notes = ns
	cmp.keysHolding = make([]bool, opts.keyCount)
	cmp.keysLowest = make([]int, opts.keyCount)
	cmp.keysAnims = make([][4]draws.Animation, opts.keyCount)
	for k := range cmp.keysAnims {
		for nt, frames := range res.framesList {
			a := draws.NewAnimation(frames, 400)
			a.SetSize(opts.kw[k], opts.H)
			a.Locate(opts.kx[k], opts.y, draws.CenterBottom)
			cmp.keysAnims[k][nt] = a
		}
	}

	cmp.h = opts.H
	cmp.keysColor = make([]color.NRGBA, opts.keyCount)
	for k := range cmp.keysColor {
		cmp.keysColor[k] = opts.Colors[opts.keyOrder[k]]
	}
	return
}

func (cmp *NotesComponent) Update(cursor float64, keysHolding []bool) {
	lowermost := cursor - game.ScreenH
	for k, idx := range cmp.keysLowest {
		var n Note
		for i := idx; i < len(cmp.notes.notes); i = n.next {
			n = cmp.notes.notes[idx]
			if n.position > lowermost {
				break
			}
			// index should be updated outside of if block.
			cmp.keysLowest[k] = i
		}

		// When Head is off the screen but Tail is on,
		// update Tail to Head since drawLongNote uses Head.
		if n.Type == Tail {
			cmp.keysLowest[k] = n.prev
		}
	}
	cmp.cursor = cursor
	cmp.keysHolding = keysHolding
}

// Notes are fixed. Lane itself moves, all notes move as same amount.
func (cmp NotesComponent) Draw(dst draws.Image) {
	uppermost := cmp.cursor + game.ScreenH
	for k, lowest := range cmp.keysLowest {
		idxs := []int{}
		var n Note
		for i := lowest; i < len(cmp.notes.notes); i = n.next {
			n = cmp.notes.notes[i]
			if n.position > uppermost {
				break
			}
			idxs = append(idxs, i)
		}

		// Make farther notes overlapped by nearer notes.
		sort.Reverse(sort.IntSlice(idxs))

		for _, i := range idxs {
			n := cmp.notes.notes[i]
			// Make long note's body overlapped by its Head and Tail.
			if n.Type == Head {
				cmp.drawLongNoteBody(dst, n)
			}

			a := cmp.keysAnims[k][n.Type]
			pos := n.position - cmp.cursor
			a.Move(0, -pos)
			if n.scored {
				// op.ColorM.ChangeHSV(0, 0.3, 0.3)
				a.ColorScale.ScaleWithColor(color.Gray{128})
			}
			a.Draw(dst)
		}
	}
}

// drawLongNoteBody draws stretched long note body sprite.
func (cmp NotesComponent) drawLongNoteBody(dst draws.Image, head Note) {
	tail := cmp.notes.notes[head.next]
	if head.Type != Head || tail.Type != Tail {
		return
	}

	a := cmp.keysAnims[head.Key][Body]
	if cmp.keysHolding[head.Key] {
		a.Reset()
	}

	length := tail.position - head.position
	length += cmp.h
	if length < 0 {
		length = 0
	}
	a.SetSize(a.W(), length)

	// Use head position because anchor is center bottom.
	pos := head.position - cmp.cursor
	a.Move(0, -pos)
	if head.scored {
		a.ColorScale.ScaleWithColor(color.Gray{128})
	}
	a.Draw(dst)
}
