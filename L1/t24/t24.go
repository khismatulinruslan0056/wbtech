package main

import (
	"fmt"
	"math"
)

func main() {
	p1 := NewPoint(10.2, 21.0)
	p2 := NewPoint(-0.2, 1.5)

	distanceP1P2 := p1.Distance(p2)
	fmt.Printf("Расстояние между точками p1 и p2 %.1f\n", distanceP1P2)
}

type Point struct {
	x float64
	y float64
}

func NewPoint(x, y float64) Point {
	return Point{
		x: x,
		y: y,
	}
}

func (p Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow(p.x-other.x, 2) + math.Pow(p.y-other.y, 2))
}
