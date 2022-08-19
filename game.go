package gosu

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hndada/gosu/audioutil"
	"github.com/hndada/gosu/ctrl"
	"github.com/hndada/gosu/db"
	"github.com/hndada/gosu/format/osr"
	"github.com/hndada/gosu/mode"
	"github.com/hndada/gosu/mode/piano"
	"github.com/hndada/gosu/render"
)

type Game struct {
	Scene
	SoundHandler ctrl.F64Handler
	SpeedHandler ctrl.F64Handler // Todo: different handler for each mode
}
type Scene interface {
	Update() any
	Draw(screen *ebiten.Image)
}

var sceneSelect *SceneSelect

func NewGame() *Game {
	mode.LoadSkin()
	piano.LoadSkin()
	db.LoadCharts(MusicPath)
	ChartInfoSprites = make([]render.Sprite, len(db.ChartInfos))

	var soundHandler ctrl.F64Handler
	var speedHandler ctrl.F64Handler
	{
		b, err := audioutil.NewBytes("skin/default-hover.wav")
		if err != nil {
			fmt.Println(err)
		}
		play := audioutil.Context.NewPlayerFromBytes(b).Play
		soundHandler = ctrl.F64Handler{
			Handler: ctrl.Handler{
				Keys:       []ebiten.Key{ebiten.KeyF1, ebiten.KeyF2},
				PlaySounds: []func(){play, play},
				HoldKey:    -1,
			},
			Min:    0,
			Max:    1,
			Unit:   0.05,
			Target: &mode.Volume,
		}
	}
	{
		b, err := audioutil.NewBytes("skin/default-hover.wav")
		if err != nil {
			fmt.Println(err)
		}
		play := audioutil.Context.NewPlayerFromBytes(b).Play
		speedHandler = ctrl.F64Handler{
			Handler: ctrl.Handler{
				Keys:       []ebiten.Key{ebiten.KeyF3, ebiten.KeyF4},
				PlaySounds: []func(){play, play},
				HoldKey:    -1,
			},
			Min:    0.1,
			Max:    2,
			Unit:   0.1,
			Target: &mode.SpeedBase,
		}
	}
	sceneSelect = NewSceneSelect()
	ebiten.SetWindowTitle("gosu")
	ebiten.SetWindowSize(WindowSizeX, WindowSizeY)
	ebiten.SetMaxTPS(mode.MaxTPS)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	g := &Game{
		Scene:        sceneSelect,
		SoundHandler: soundHandler,
		SpeedHandler: speedHandler,
	}
	return g
}

type SelectToPlayArgs struct {
	Path   string
	Mode   int
	Replay *osr.Format
	Play   bool
}

func (g *Game) Update() error {
	args := g.Scene.Update()
	if args == nil {
		return nil
	}
	switch args := args.(type) {
	case mode.PlayToResultArgs:
		// Todo: selectResult
		g.Scene = sceneSelect
	case SelectToPlayArgs:
		switch args.Mode {
		// case args.Mode&mode.ModePiano != 0:
		case mode.ModePiano4, mode.ModePiano7:
			var err error
			g.Scene, err = piano.NewScenePlay(args.Path, args.Replay, args.Play)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	g.Scene.Draw(screen)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenSizeX, screenSizeY
}

//	func a(args *Args) {
//		args2 := reflect.ValueOf(args)
//
// from := args2.FieldByName("From").String()
//
//		NewSceneResult()
//		args2.FieldByName("Result")
//	}
