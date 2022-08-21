package main

import "math"

type Vector2 struct {
	x float64
	y float64
}

func (v *Vector2) PUnit() {
	dist := v.Distance(Vector2{0, 0})
	v.x /= dist
	v.y /= dist
}

func (v Vector2) Unit() Vector2 {
	dist := v.Distance(Vector2{0, 0})

	return Vector2{v.x / dist, v.y / dist}
}

func (v Vector2) Sub(u Vector2) Vector2 {
	return Vector2{u.x - v.x, u.y - v.y}
}

func (v Vector2) Dot(u Vector2) float64 {
	return v.x*u.x + v.y*u.y
}

func (v Vector2) Distance(u Vector2) float64 {
	return math.Sqrt(math.Pow((u.x-v.x), 2) + math.Pow((u.y-v.y), 2))
}
