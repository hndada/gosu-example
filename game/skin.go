package game

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
)

// skin -> spritesheet
// 마지막으로 불러온 스킨 불러오기: 처음 / 오류 발생 시 defaultSkin
const (
	ScoreComma = iota + 10
	ScoreDot
	ScorePercent
)

// image.Image가 아닌 *ebiten.Image로 해야 이미지 자체가 한 번만 로드 됨
var Skin struct {
	Number [13]*ebiten.Image // including dot, comma, percent
	// Combo     [10]*ebiten.Image
	BoxLeft   *ebiten.Image
	BoxRight  *ebiten.Image
	BoxMiddle *ebiten.Image

	Cursor      *ebiten.Image
	CursorSmoke *ebiten.Image

	DefaultBG *ebiten.Image
}

func LoadImage(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	ei, _ := ebiten.NewImageFromImage(i, ebiten.FilterDefault)
	return ei, nil
}

func LoadSkin(cwd string) {
	var err error
	dir := filepath.Join(cwd, "skin")
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("score-%d.png", i)
		path := filepath.Join(dir, name)
		Skin.Number[i], err = LoadImage(path)
		if err != nil {
			log.Fatal(err)
		}
	}
}
