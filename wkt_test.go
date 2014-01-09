// Copyright 2014 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wkt

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		data string
		geom Geom
	}{
		{
			`POINT(-7.9270020000000070 71.1508198000000505)`,
			&Point{Coord{-7.9270020000000070, 71.1508198000000505, 0, 0}, 0},
		},
		{
			`POINT ZM(1.0 2.0 3.0 4.0)`,
			&Point{Coord{1.0, 2.0, 3.0, 4.0}, ZM},
		},
		{
			`point z(1.0 2.0 3.0)`,
			&Point{Coord{1.0, 2.0, 3.0, 0}, Z},
		},
		{
			`point m(1.0 2.0 4.0)`,
			&Point{Coord{1.0, 2.0, 0, 4.0}, M},
		},
		{
			`MULTIPOINT(30 10, 10 30, 40 40)`,
			&MultiPoint{Coords: []Coord{{30, 10, 0, 0}, {10, 30, 0, 0}, {40, 40, 0, 0}}},
		},
		{
			`MULTIPOINT((30 10), (10 30), (40 40))`,
			&MultiPoint{Coords: []Coord{{30, 10, 0, 0}, {10, 30, 0, 0}, {40, 40, 0, 0}}},
		},
		{
			`LINESTRING(30 10, 10 30, 40 40)`,
			&LineString{Coords: []Coord{{30, 10, 0, 0}, {10, 30, 0, 0}, {40, 40, 0, 0}}},
		},
		{
			`POLYGON((30 10, 10 20, 20 40, 40 40, 30 10))`,
			&Polygon{Rings: [][]Coord{
				{{30, 10, 0, 0}, {10, 20, 0, 0}, {20, 40, 0, 0}, {40, 40, 0, 0}, {30, 10, 0, 0}},
			}},
		},
		{
			`POLYGON M((35 10 1, 10 20 2, 15 40 3, 45 45 4, 35 10 1),(20 30 1, 35 35 2, 30 20 3, 20 30 1))`,
			&Polygon{Rings: [][]Coord{
				{{35, 10, 0, 1}, {10, 20, 0, 2}, {15, 40, 0, 3}, {45, 45, 0, 4}, {35, 10, 0, 1}},
				{{20, 30, 0, 1}, {35, 35, 0, 2}, {30, 20, 0, 3}, {20, 30, 0, 1}},
			}, Opt: M},
		},
		{
			`MULTIPOLYGON(
			((30 10, 10 20, 20 40, 40 40, 30 10)),
			((35 10, 10 20, 15 40, 45 45, 35 10),(20 30, 35 35, 30 20, 20 30))
		)`,
			&MultiPolygon{Polygons: [][][]Coord{
				{
					{{30, 10, 0, 0}, {10, 20, 0, 0}, {20, 40, 0, 0}, {40, 40, 0, 0}, {30, 10, 0, 0}},
				},
				{
					{{35, 10, 0, 0}, {10, 20, 0, 0}, {15, 40, 0, 0}, {45, 45, 0, 0}, {35, 10, 0, 0}},
					{{20, 30, 0, 0}, {35, 35, 0, 0}, {30, 20, 0, 0}, {20, 30, 0, 0}},
				},
			}},
		},
	}
	for i, test := range tests {
		g, err := Parse([]byte(test.data))
		if err != nil {
			t.Errorf("test %d: %v", i, err)
			continue
		}
		if !g.Equal(test.geom) {
			t.Errorf("test %d: expected %+v\ngot\t %+v", i, test.geom, g)
		}
	}
}

func TestParseErrors(t *testing.T) {
	// lets cover 100%
	tests := [][]byte{
		nil,
		[]byte("P"),
		[]byte("POINT("),
		[]byte("POINT#"),
		[]byte("POINT(0 0,0 0)"),
		[]byte("MULTIPOINT(0 0 #"),
		[]byte("MULTIPOINT((0 0,"),
		[]byte("MULTIPOINT((0 0)#"),
		[]byte("MULTIPOINT((0 0), #"),
		[]byte("POLYGON#"),
		[]byte("POLYGON(#"),
		[]byte("POLYGON((30 10, 10 20, 20 40))"),
		[]byte("POLYGON((30 10, 10 20, 20 40, 30 11))"),
		[]byte("POLYGON((30 10, 10 20, 20 40, 30 10)#"),
		[]byte("MULTIPOLYGON#"),
		[]byte("MULTIPOLYGON(#"),
		[]byte("MULTIPOLYGON(((30 10, 10 20, 20 40, 30 10))#"),
	}
	for _, test := range tests {
		g, err := Parse(test)
		if err == nil {
			t.Error("expected error got", g)
		} else {
			t.Log(err)
		}
	}
}

func TestOpt(t *testing.T) {
	tests := []struct {
		data string
		z    bool
		m    bool
	}{
		{"point(1 2 3 4)", false, false}, // not strict
		{"point z(1 2 3)", true, false},
		{"point m(1 2 4)", false, true},
		{"point zm(1 2 3 4)", true, true},
	}
	for i, test := range tests {
		g, err := Parse([]byte(test.data))
		if err != nil {
			t.Errorf("test %d: %v", i, err)
			continue
		}
		if test.z != g.Is3d() {
			t.Errorf("test %d: expected 3d to be %v", i, test.z)
		}
		if test.m != g.IsMeasured() {
			t.Errorf("test %d: expected measured to be %v", i, test.m)
		}
	}
}
