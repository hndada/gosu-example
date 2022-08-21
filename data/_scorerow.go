package data

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/hndada/gosu/mode"
	"github.com/vmihailenco/msgpack/v5"
)

type ScoreRow struct {
	MD5  [md5.Size]byte
	Time time.Time
	// Key log (Replay)
	mode.Result
}

var Scores = make(map[[md5.Size]byte][]ScoreRow)

func LoadScores() {
	const fname = "score.db"
	// for i, ci := range s.View {
	// 	d, err := os.ReadFile(ci.Path)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	s.MD5ToIndexMap[md5.Sum(d)] = i
	// 	s.IndexToMD5Map[i] = md5.Sum(d)
	// }
	b, err := os.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		os.Rename("score.db", "score_crashed.db") // Fail if not existed.
	}
	// r := bytes.NewReader(b)
	msgpack.Unmarshal(b, &Scores)
}
func SaveScores() {
	b, err := msgpack.Marshal(&Scores)
	if err != nil {
		fmt.Printf("Failed to save error: %s", err)
		return
	}
	err = os.WriteFile("score.db", b, 0644)
	if err != nil {
		fmt.Printf("Failed to save error: %s", err)
		return
	}
}

// Todo: implement it
// func ReplayScore(c *Chart, rf *osr.Format) int {
// 	s := NewScenePlay(c, "", rf, false)
// 	for !s.IsFinished() {
// 		s.Update(nil)
// 	}
// 	return int(s.Score())
// }

// fs, err := os.ReadDir("replay")
// if err != nil {
// 	panic(err)
// }
// for _, f := range fs {
// 	if f.IsDir() || filepath.Ext(f.Name()) != ".osr" {
// 		continue
// 	}
// 	rd, err := os.ReadFile(filepath.Join("replay", f.Name()))
// 	if err != nil {
// 		panic(err)
// 	}
// 	rf, err := osr.Parse(rd)
// 	if err != nil {
// 		panic(err)
// 	}
// 	s.Replays = append(s.Replays, rf)
// }

// // FetchReplay returns first MD5-matching replay format.
// func (s SceneSelect) FetchReplay(i int) *osr.Format {
// 	md5 := s.IndexToMD5Map[i]
// 	for _, r := range s.Replays {
// 		if md5 == r.MD5() {
// 			return r
// 		}
// 	}
// 	return nil
// }