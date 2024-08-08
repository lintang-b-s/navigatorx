package alg

import (
	"fmt"

	"github.com/dhconnelly/rtreego"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// sebelum pake Uber h3 buat spatial index, project ini pake R-tree

var tol = 0.0001

type StreetRect struct {
	Location rtreego.Point
	Wormhole chan int
	Street   *SurakartaWay
}

func (s *StreetRect) Bounds() rtreego.Rect {
	// define the bounds of s to be a rectangle centered at s.location
	// with side lengths 2 * tol:  (https://github.com/dhconnelly/rtreego?tab=readme-ov-file#documentation)
	return s.Location.ToRect(tol)
}

// rtree
type Rtree struct {
	StRtree *rtreego.Rtree
}

func NewRtree(stR *rtreego.Rtree) *Rtree {
	return &Rtree{
		stR,
	}
}

// gak bisa simpen rtreenya ke file binary (udah coba)
func BikinRtreeStreetNetwork(ways []SurakartaWay, ch *ContractedGraph, nodeIdxMap map[int64]int32) {
	bar := progressbar.NewOptions(len(ways),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][4/7][reset] Membuat rtree entry dari osm way/edge ..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	rtg := rtreego.NewTree(2, 25, 50) // 2 dimension, 25 min entries dan 50 max entries
	rt := NewRtree(rtg)
	for _, way := range ways {
		rt.StRtree.Insert(&StreetRect{Location: rtreego.Point{way.CenterLoc[0], way.CenterLoc[1]},
			Wormhole: nil,
			Street:   &way})
		bar.Add(1)
	}
	fmt.Println("")
	ch.Rtree = rt
}
