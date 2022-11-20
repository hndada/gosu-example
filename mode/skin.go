package mode

import (
	"fmt"
	"io/fs"

	"github.com/hndada/gosu/draws"
)

// In narrow meaning, Skin stands for a set of Sprites.
// In wide meaning, Skin also includes a set of sounds.
// Package defaultskin has a set of sounds.
type Skin struct {
	Type              int
	DefaultBackground draws.Sprite
	Score             [13]draws.Sprite // number + sign(. , %)
}

const (
	ScoreDot = iota
	ScoreComma
	ScorePercent
)

var (
	DefaultSkin = Skin{Type: Default}
	UserSkin    = Skin{Type: User}
	// PlaySkin    = Skin{Type: Play}
)

// s stands for sprite.
// a stands for animation.
// S stands for UserSettings.
func (skin *Skin) Load(fsys fs.FS) {
	S := UserSettings // abbreviation
	if skin.Type == Play {
		skin.Reset()
	}
	skin.DefaultBackground = NewBackground(fsys, "interface/default-bg.jpg")
	for i := 0; i < 10; i++ {
		s := draws.NewSprite(fsys, fmt.Sprintf("score/%d.png", i))
		s.ApplyScale(S.ScoreScale)
		// Need to set same base line, since each number has different height.
		if i == 0 {
			s.Locate(ScreenSizeX, 0, draws.RightTop)
		} else {
			s.Locate(ScreenSizeX, skin.Score[0].H()-s.H(), draws.RightTop)
		}
		skin.Score[i] = s
	}
	for i, name := range []string{"dot", "comma", "percent"} {
		s := draws.NewSprite(fsys, fmt.Sprintf("score/%s.png", name))
		s.ApplyScale(S.ScoreScale)
		s.Locate(ScreenSizeX, 0, draws.RightTop)
		skin.Score[10+i] = s
	}
	base := []Skin{{}, DefaultSkin, UserSkin}[skin.Type]
	skin.fillBlank(base)
}
func (skin *Skin) fillBlank(base Skin) {
	if !skin.DefaultBackground.IsValid() {
		skin.DefaultBackground = base.DefaultBackground
	}
	for _, s := range skin.Score {
		if !s.IsValid() {
			skin.Score = base.Score
			break
		}
	}
}
func (skin *Skin) Reset() {
	kind := skin.Type
	switch kind {
	case User:
		*skin = DefaultSkin
	case Play:
		*skin = UserSkin
	}
	skin.Type = kind
}

func NewBackground(fsys fs.FS, name string) draws.Sprite {
	s := draws.NewSprite(fsys, name)
	s.ApplyScale(ScreenSizeX / s.W())
	s.Locate(ScreenSizeX/2, ScreenSizeY/2, draws.CenterMiddle)
	return s
}
