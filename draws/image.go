package draws

import (
	"image"
	"image/color"
	"io/fs"
	"net/http"

	// Following imports are required.
	_ "image/jpeg"
	_ "image/png"

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
// ebiten.NewImageFromImage will panic when input is nil.
func LoadImage(fsys fs.FS, name string) Image {
	if i := LoadImageImage(fsys, name); i != nil {
		return Image{ebiten.NewImageFromImage(i)}
	}
	return Image{}
}
func LoadImageFromURL(url string) (Image, error) {
	res, err := http.Get(url)
	if err != nil {
		return Image{}, err
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	if err != nil {
		return Image{}, err
	}
	eimg := ebiten.NewImageFromImage(img)
	return Image{eimg}, err
}

// LoadImageImage returns image.Image.
func LoadImageImage(fsys fs.FS, name string) image.Image {
	f, err := fsys.Open(name)
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
