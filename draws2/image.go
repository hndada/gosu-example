package draws

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewImage returns nil when fails to load image from the path.
func NewImage(path string) *ebiten.Image {
	if i := NewImageImage(path); i != nil {
		return ebiten.NewImageFromImage(i)
	}
	return nil
}

// NewImageImage returns image.Image.
func NewImageImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return src
}
func NewImages(path string) (is []*ebiten.Image) {
	const ext = ".png"
	one := []*ebiten.Image{NewImage(path + ext)}
	dir, err := os.Open(path)
	if err != nil {
		return one
	}
	defer dir.Close()
	fs, err := dir.ReadDir(-1)
	if err != nil {
		return one
	}

	nums := make([]int, 0, len(fs))
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		num := strings.TrimSuffix(f.Name(), ext)
		if num, err := strconv.Atoi(num); err == nil {
			nums = append(nums, num)
		}
	}
	sort.Ints(nums)
	for _, num := range nums {
		path := filepath.Join(path, fmt.Sprintf("%d.png", num))
		is = append(is, NewImage(path))
	}
	return
}

func NewXFlippedImage(i *ebiten.Image) *ebiten.Image {
	w, h := i.Size()
	i2 := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(w), 0)
	i2.DrawImage(i, op)
	return i2
}
func NewYFlippedImage(i *ebiten.Image) *ebiten.Image {
	w, h := i.Size()
	i2 := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, -1)
	op.GeoM.Translate(0, float64(h))
	i2.DrawImage(i, op)
	return i2
}
func NewScaledImage(i *ebiten.Image, scale float64) *ebiten.Image {
	sw, sh := i.Size()
	w, h := math.Ceil(float64(sw)*scale), math.Ceil(float64(sh)*scale)
	i2 := ebiten.NewImage(int(w), int(h))
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(scale, scale)
	i2.DrawImage(i, op)
	return i2
}
func ImageSize(i *ebiten.Image) Vector2 { return IntVec2(i.Size()) }
