package alg

import (
	"github.com/dhconnelly/rtreego"
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
