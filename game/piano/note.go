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

type NoteKind int

const (
	Normal NoteKind = iota
	Head
	Tail
	Body
)

type Note struct {
	Time   int32
	Kind   NoteKind
	Key    int
	Sample game.Sample

	position float64 // Scaled x or y value.
	next     int     // For updating staged notes.
	prev     int     // For accessing to Head from Tail.
	scored   bool
}

// The length of the returned slice is 1 or 2.
func newNoteFromOsu(f osu.HitObject, keyCount int) (ns []Note) {
	n := Note{
		Time:   int32(f.Time),
		Kind:   Normal,
		Key:    f.Column(keyCount),
		Sample: game.NewSample(f),
	}
	if f.NoteType&osu.ComboMask == osu.HitTypeHoldNote {
		n.Kind = Head
		d := int32(f.EndTime) - n.Time
		n2 := Note{
			Time: n.Time + d,
			Kind: Tail,
			Key:  n.Key,
			// Tail has no sample sound.
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
		if n.Kind == Tail {
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
	kfni := make([]int, keyCount)
	for k := range kfni {
		kfni[k] = len(ns)
	}
	for k := range kfni {
		for i, n := range ns {
			if k == n.Key {
				kfni[n.Key] = i
				break
			}
		}
	}

	return Notes{ns, keyCount, kfni, nil}
}

func (ns Notes) NoteCounts() []int {
	counts := make([]int, 2)
	for _, n := range ns.notes {
		switch n.Kind {
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
	keyCount   int
	keyOrder   []KeyKind
	keysW      []float64
	H          float64 // Applies to all types of notes.
	keysX      []float64
	y          float64 // center bottom
	TailOffset int32
	Colors     [4]color.NRGBA
	// LongBodyStyle     int // Stretch or Attach.
}

func NewNotesOptions(stage StageOptions, keys KeysOptions) NotesOptions {
	return NotesOptions{
		keyCount:   stage.keyCount,
		keyOrder:   keys.Order(),
		keysW:      keys.w,
		H:          20,
		keysX:      keys.x,
		y:          keys.y,
		TailOffset: 0,
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
	keysAnims [][4]draws.Animation
	Notes
	keysLowest  []int // indexes of lowest notes
	cursor      float64
	keysColor   []color.NRGBA
	keysHolding []bool
	// h           float64 // used for drawLongNoteBody
}

func NewNotesComponent(res NotesResources, opts NotesOptions, ns Notes, dys game.Dynamics) (cmp NotesComponent) {
	cmp.keyCount = opts.keyCount
	cmp.keysAnims = make([][4]draws.Animation, opts.keyCount)
	for k := range cmp.keysAnims {
		for nk, frames := range res.framesList {
			a := draws.NewAnimation(frames, 400)
			a.SetSize(opts.keysW[k], opts.H)
			if nk == int(Body) {
				a.Locate(opts.keysX[k], opts.y, draws.CenterTop)
			} else {
				a.Locate(opts.keysX[k], opts.y, draws.CenterBottom)
			}
			cmp.keysAnims[k][nk] = a
		}
	}

	for i, n := range ns.notes {
		if n.Kind != Tail {
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
	cmp.Notes = ns
	cmp.keysLowest = make([]int, opts.keyCount)

	cmp.keysColor = make([]color.NRGBA, opts.keyCount)
	for k := range cmp.keysColor {
		cmp.keysColor[k] = opts.Colors[opts.keyOrder[k]]
	}
	cmp.keysHolding = make([]bool, opts.keyCount)
	// cmp.h = opts.H
	return
}

func (cmp *NotesComponent) keysFocusNote() []Note {
	ns := make([]Note, cmp.keyCount)
	for k, ni := range cmp.keysFocus {
		if ni == len(cmp.notes) {
			continue
		}
		ns[k] = cmp.notes[ni]
	}
	return ns
}

func (cmp *NotesComponent) Update(ka game.KeyboardAction, cursor float64) {
	lowermost := cursor - game.ScreenH
	for k, lowest := range cmp.keysLowest {
		for ni := lowest; ni < len(cmp.notes); ni = cmp.notes[ni].next {
			n := cmp.notes[lowest]
			if n.position > lowermost {
				break
			}
			// index should be updated outside of if block.
			cmp.keysLowest[k] = ni
		}
		// When Head is off the screen but Tail is on,
		// update Tail to Head since drawLongNote uses Head.
		ni := cmp.keysLowest[k]
		if n := cmp.notes[ni]; n.Kind == Tail {
			cmp.keysLowest[k] = n.prev
		}
	}
	cmp.cursor = cursor
	cmp.keysHolding = ka.KeysHolding()
}

// Notes are fixed. Lane itself moves, all notes move as same amount.
func (cmp NotesComponent) Draw(dst draws.Image) {
	uppermost := cmp.cursor + game.ScreenH
	for k, lowest := range cmp.keysLowest {
		var nis []int
		for ni := lowest; ni < len(cmp.notes); ni = cmp.notes[ni].next {
			n := cmp.notes[ni]
			if n.position > uppermost {
				break
			}
			nis = append(nis, ni)
		}

		// Make farther notes overlapped by nearer notes.
		sort.Sort(sort.Reverse(sort.IntSlice(nis)))

		for _, ni := range nis {
			n := cmp.notes[ni]
			// Make long note's body overlapped by its Head and Tail.
			if n.Kind == Head {
				cmp.drawLongNoteBody(dst, n)
			}

			a := cmp.keysAnims[k][n.Kind]
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
	tail := cmp.notes[head.next]
	if head.Kind != Head || tail.Kind != Tail {
		return
	}

	a := cmp.keysAnims[head.Key][Body]
	if !cmp.keysHolding[head.Key] {
		a.Reset()
	}

	length := tail.position - head.position
	// length += cmp.h
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
