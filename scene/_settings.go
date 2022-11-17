package scene

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

var (
	musicRoots   []string
	windowSizeX  int
	windowSizeY  int
	musicVolume  float64
	effectVolume float64
	cursorScale  float64
)

type Settings struct {
	TPS         int
	ScreenSizeX int
	ScreenSizeY int

	MusicRoots  []string
	WindowSizeX int
	WindowSizeY int
	VolumeMusic float64
	VolumeSound float64
	CursorScale float64
}

// Default settings should not be directly exported.
// It may be modified by others.
func (Settings) Default() Setter {
	return Settings{
		TPS:         TPS,
		ScreenSizeX: ScreenSizeX,
		ScreenSizeY: ScreenSizeY,

		MusicRoots:  []string{"music"},
		WindowSizeX: 1600,
		WindowSizeY: 900,
		VolumeMusic: 0.25,
		VolumeSound: 0.25,
		CursorScale: 0.1,
	}
}

// generated by settings.PrintCurrent
func (Settings) Current() Setter {
	return Settings{
		TPS:         TPS,
		ScreenSizeX: ScreenSizeX,
		ScreenSizeY: ScreenSizeY,

		MusicRoots:  musicRoots,
		WindowSizeX: windowSizeX,
		WindowSizeY: windowSizeY,
		VolumeMusic: musicVolume,
		VolumeSound: effectVolume,
		CursorScale: cursorScale,
	}
}

// generated by settings.PrintSet
func (Settings) Set(s Setter) {
	s.(Settings).set(s.(Settings))
}
func (Settings) set(s Settings) {
	musicRoots = s.MusicRoots
	windowSizeX = s.WindowSizeX
	windowSizeY = s.WindowSizeY
	musicVolume = s.VolumeMusic
	effectVolume = s.VolumeSound
	cursorScale = s.CursorScale
}