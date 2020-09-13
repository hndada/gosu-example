package mania

import (
	"github.com/hndada/gosu/mode"
	"github.com/hndada/rg-parser/osugame/osu"
)

type Chart struct {
	*mode.BaseChart
	Keys  int
	Notes []Note
}

// raw 차트에는 모드가 들어가면 안됨
// 모드마다 TransPoint(TimingPoint), Note건듦
func NewChartFromOsu(o *osu.Format, path string) (*Chart, error) {
	var c Chart
	c.BaseChart = mode.NewBaseChartFromOsu(o, path)
	c.Keys = int(c.Parameter["Scale"])
	err := c.loadNotes(o)
	if err != nil {
		panic(err)
	}
	// c.CalcLevel() // 노트별 strain 계산 후 차트에 level 입력
	// c.CalcScore()
	return &c, nil
}

func (c *Chart) ApplyMods(mods Mods) *Chart {
	var c2 Chart
	c2.BaseChart = c.BaseChart // todo: pointer?
	c2.Keys = c.Keys
	c2.Notes = make([]Note, len(c.Notes))
	for i, n := range c.Notes {
		n.Time = int64(float64(n.Time) / mods.TimeRate)
		n.Time2 = int64(float64(n.Time2) / mods.TimeRate)
		if mods.Mirror { // todo: scartch는 따로 분리? -> 까다로워질지도, 아니면 미러로 그냥 쇼부 봐
			n.Key = c.Keys - 1 - n.Key
		}
		c2.Notes[i] = n
	}
	return &c2
}

func (c Chart) EndTime() int64 {
	return c.Notes[len(c.Notes)-1].Time
}
func (c Chart) LNCount() int {
	var num int
	for _, n := range c.Notes {
		if n.Type == TypeLNHead {
			num++
		}
	}
	return num
}
