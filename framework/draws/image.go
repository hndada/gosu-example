package draws

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type Image struct{ *ebiten.Image }

func (i Image) IsValid() bool { return i.Image != nil }
func (i Image) Size() Vector2 {
	if !i.IsValid() {
		return Vector2{}
	}
	return IntVec2(i.Image.Size())
}
func (i Image) Draw(dst Image, op Op) {
	dst.Image.DrawImage(i.Image, &op)
}

func NewImage(w, h float64) Image {
	return Image{ebiten.NewImage(int(w), int(h))}
}

// LoadImage returns nil when fails to load image from the path.
func LoadImage(path string) Image {
	// ebiten.NewImageFromImage will panic when input is nil.
	if i := LoadImageImage(path); i != nil {
		return Image{ebiten.NewImageFromImage(i)}
	}
	return Image{}
}

// LoadImageImage returns image.Image.
func LoadImageImage(path string) image.Image {
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

func LoadImages(path string) (is []Image) {
	const ext = ".png"
	one := []Image{LoadImage(path + ext)}
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
		is = append(is, LoadImage(path))
	}
	return
}
func NewImageXFlipped(src Image) Image {
	size := src.Size()
	dst := Image{ebiten.NewImage(size.XYInt())}
	op := Op{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(size.X, 0)
	src.Draw(dst, op)
	return dst
}
func NewImageYFlipped(src Image) Image {
	size := src.Size()
	dst := Image{ebiten.NewImage(size.XYInt())}
	op := Op{}
	op.GeoM.Scale(1, -1)
	op.GeoM.Translate(0, size.Y)
	src.Draw(dst, op)
	return dst
}
func NewImageColored(src Image, color color.Color) Image {
	size := src.Size()
	dst := Image{ebiten.NewImage(size.XYInt())}
	op := Op{}
	op.ColorM.ScaleWithColor(color)
	src.Draw(dst, op)
	return dst
}

//	func NewImageScaled(src Image, scale float64) Image {
//		size := src.Size().Mul(Scalar(scale))
//		dst := Image{ebiten.NewImage(size.XYInt())}
//		op := Op{}
//		op.GeoM.Scale(scale, scale)
//		op.GeoM.Translate(0, size.Y)
//		src.Draw(dst, op)
//		return dst
//	}