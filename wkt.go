// Copyright 2014 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package wkt implements a simplified parser for Well Known Text
// http://en.wikipedia.org/wiki/Well-known_text
package wkt

// Geom is one of the supported geometry types supported by this package
type Geom interface {
	IsMeasured() bool
	Is3d() bool
	Equal(Geom) bool
}

// Parse returns the parsed geometry or an error
func Parse(data []byte) (Geom, error) {
	s := &scanner{raw: data}
	return s.scanGeom()
}

// Indicators for declared use of either z or m coordinates or both
const (
	_ Opt = iota
	Z
	M
	ZM
)

// Opt indicates which additional coordinate options are used for this geometry
type Opt uint
// Is3d returns whether the z-coordinate is declared to be used as in "POINT Z(1 2 3)" or "POINT ZM(1 2 3 4)"
func (o Opt) Is3d() bool       { return o&Z != 0 }
// IsMeasured returns whether the m-coordinate is declared to be used as in "POINT M(1 2 4)" or "POINT ZM(1 2 3 4)"
func (o Opt) IsMeasured() bool { return o&M != 0 }

// Coord represents a single location in a coordinate space
type Coord struct {
	X, Y, Z, M float64
}

// Point reprsents a point in 2 or 3 dimensions with or without measure
type Point struct {
	Coord
	Opt
}

// Equal returns whether this point equals g
func (p *Point) Equal(g Geom) bool {
	o, ok := g.(*Point)
	return ok && *p == *o
}

func equalCoords(c, o []Coord) bool {
	if len(c) != len(o) {
		return false
	}
	for i, cc := range c {
		if cc != o[i] {
			return false
		}
	}
	return true
}

// MultiPoint is a list unconnected points
type MultiPoint struct {
	Coords []Coord
	Opt
}

// Equal returns whether this multipoint equals g
func (m *MultiPoint) Equal(g Geom) bool {
	o, ok := g.(*MultiPoint)
	return ok && m.Opt == o.Opt && equalCoords(m.Coords, o.Coords)
}

// LineString is a list of connected points
type LineString struct {
	Coords []Coord
	Opt
}

// Equal returns whether this linestring equals g
func (l *LineString) Equal(g Geom) bool {
	o, ok := g.(*LineString)
	return ok && l.Opt == o.Opt && equalCoords(l.Coords, o.Coords)
}

// Polygon is a list of rings where the first ring is the exterior (outline)
// and following are interior rings (holes)
type Polygon struct {
	Rings [][]Coord
	Opt
}

// Equal returns whether this polygon equals g
func (p *Polygon) Equal(g Geom) bool {
	o, ok := g.(*Polygon)
	if !ok {
		return false
	}
	if p.Opt != o.Opt {
		return false
	}
	return equalRings(p.Rings, o.Rings)
}

func equalRings(r, o [][]Coord) bool {
	if len(r) != len(o) {
		return false
	}
	for i, rc := range r {
		if !equalCoords(rc, o[i]) {
			return false
		}
	}
	return true
}

// MultiPolygon is a list of multiple polygon ring lists
type MultiPolygon struct {
	Polygons [][][]Coord
	Opt
}

// Equal returns whether this multipolygon equals g
func (m *MultiPolygon) Equal(g Geom) bool {
	o, ok := g.(*MultiPolygon)
	if !ok {
		return false
	}
	if m.Opt != o.Opt {
		return false
	}
	if len(m.Polygons) != len(o.Polygons) {
		return false
	}
	for i, p := range m.Polygons {
		if !equalRings(p, o.Polygons[i]) {
			return false
		}
	}
	return true
}
