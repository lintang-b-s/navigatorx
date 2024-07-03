package alg

import (
	"github.com/dhconnelly/rtreego"
)

var tol = 0.0001

type StreetRect struct {
	Location rtreego.Point
	Wormhole chan int
	Street   *SurakartaWay
}

func (s *StreetRect) Bounds() rtreego.Rect {
	// define the bounds of s to be a rectangle centered at s.location
	// with side lengths 2 * tol:
	return s.Location.ToRect(tol)
}

// rtree
var StRTree = rtreego.NewTree(2, 25, 50) // 2 dimension, 25 min entries dan 50 max entries
