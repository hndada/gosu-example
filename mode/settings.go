package mode

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/hndada/gosu/defaultskin"
	"github.com/hndada/gosu/input"
)

const (
	// Currently, TPS should be 1000 or greater.
	// TPS supposed to be multiple of 1000, since only one speed value
	// goes passed per Update, while unit of TransPoint's time is 1ms.
	// TPS affects only on Update(), not on Draw().
	// Todo: add lower TPS support
	TPS = 1000

	// ScreenSize is a logical size of in-game screen.
	ScreenSizeX = 1600
	ScreenSizeY = 900
)

// Struct as a type of settings value is discouraged.
// Unmatched fields will not be touched, feel free to pre-fill default values.
// Todo: alert warning message to user when some lines are failed to be decoded
type Settings struct {
	VolumeMusic          float64
	VolumeSound          float64
	Offset               int64
	BackgroundBrightness float64

	MeterUnit     float64 // number of pixels per 1ms
	MeterHeight   float64
	ScoreScale    float64
	ScoreDigitGap float64
}

const (
	Default = iota
	User
	Play // For skin. Refreshes on every play.
)

func NewSettings() Settings {
	return Settings{
		VolumeMusic:          0.25,
		VolumeSound:          0.25,
		Offset:               -65,
		BackgroundBrightness: 0.6,

		MeterUnit:     4,
		MeterHeight:   50,
		ScoreScale:    0.65,
		ScoreDigitGap: 0,
	}
}

var (
	DefaultSettings = NewSettings()
	UserSettings    = NewSettings()
	S               = &UserSettings
)

func init() {
	DefaultSettings.process()
	UserSettings.process()
	DefaultSkin.Load(defaultskin.FS)
	UserSkin.Load(defaultskin.FS)
}

func (settings *Settings) Load(data string) {
	*settings = NewSettings()
	_, err := toml.Decode(data, settings)
	if err != nil {
		fmt.Println(err)
	}
	defer settings.process()

	Normalize(&settings.VolumeMusic, 0, 1)
	Normalize(&settings.VolumeSound, 0, 1)
	Normalize(&settings.Offset, -300, 300)
	Normalize(&settings.BackgroundBrightness, 0, 1)

	Normalize(&settings.MeterUnit, 0, 5)
	Normalize(&settings.MeterHeight, 0, 100)
	Normalize(&settings.ScoreScale, 0, 2)
	Normalize(&settings.ScoreDigitGap, -0.05, 0.05)
}

// process processes from raw settings to screen size-adjusted settings.
func (settings *Settings) process() {
	settings.ScoreDigitGap *= ScreenSizeX
}

type Number interface{ int | int64 | float64 }

func Normalize[V Number](v *V, min, max V) {
	if *v > max {
		*v = max
	}
	if *v < min {
		*v = min
	}
}
func NormalizeKeys(names []string) {
	keys := input.NamesToKeys(names)
	m := make(map[input.Key]bool)
	for i, k := range keys {
		if m[k] {
			fmt.Printf("some keys are duplicated: %v\n", names)
			keys[i] = input.KeyNone
		}
		m[k] = true
	}
}
