// Copyright 2014 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wkt

import (
	"bytes"
	"fmt"
	"io"
)

type scanner struct {
	raw []byte
	i   int
	opt Opt
}

func (s *scanner) peek() (byte, error) {
	if s.i >= len(s.raw) {
		return '\x00', io.ErrUnexpectedEOF
	}
	return s.raw[s.i], nil
}

func (s *scanner) skipWs() {
	if s.i >= len(s.raw) {
		return
	}
	for i, b := range s.raw[s.i:] {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		}
		s.i += i
		return
	}
}
func (s *scanner) scanStart() error {
	s.skipWs()
	c, err := s.peek()
	if err != nil {
		return err
	}
	if c != '(' {
		return fmt.Errorf("expect '(' got %q", c)
	}
	s.i++
	return nil
}

func (s *scanner) scanContinue() (bool, error) {
	s.skipWs()
	c, err := s.peek()
	if err != nil {
		return false, err
	}
	comma := c == ','
	if !comma && c != ')' {
		return false, fmt.Errorf("expect ',' or ')' got %q", c)
	}
	s.i++
	return comma, nil
}

func (s *scanner) scanIdent() (string, error) {
	s.skipWs()
	if s.i >= len(s.raw) {
		return "", io.ErrUnexpectedEOF
	}
	var (
		ident []byte
		b     byte
		i     int
	)
	for i, b = range s.raw[s.i:] {
		lower := b >= 'a' && b <= 'z'
		if lower || b >= 'A' && b <= 'Z' {
			if lower {
				b = b - 'a' + 'A'
			}
			ident = append(ident, b)
			continue
		}
		break
	}
	if len(ident) == 0 {
		return "", fmt.Errorf("no ident byte %q", b)
	}
	s.i += i
	return string(ident), nil
}
func (s *scanner) scanCoord() (c Coord, comma bool, err error) {
	s.skipWs()
	if s.i >= len(s.raw) {
		return c, false, io.ErrUnexpectedEOF
	}
	r := bytes.NewReader(s.raw[s.i:])
	var fs []*float64
	if s.opt == M {
		fs = []*float64{&c.X, &c.Y, &c.M}
	} else {
		fs = []*float64{&c.X, &c.Y, &c.Z, &c.M}
	}
	for _, f := range fs {
		_, err := fmt.Fscan(r, f)
		if err != nil {
			return c, false, err
		}
		s.i = len(s.raw) - r.Len()
		s.skipWs()
		b, err := s.peek()
		if err != nil {
			return c, false, io.ErrUnexpectedEOF
		}
		if comma = b == ','; comma || b == ')' {
			s.i++
			break
		}
	}
	return
}
func (s *scanner) scanCoords(multi bool) ([]Coord, error) {
	err := s.scanStart()
	if err != nil {
		return nil, err
	}
	var cs []Coord
	var c Coord
	var comma bool
	if multi {
		err = s.scanStart()
		multi = err == nil
	}
	for {
		c, comma, err = s.scanCoord()
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
		if comma {
			if multi {
				return nil, fmt.Errorf("expect ')' got ','")
			}
			continue
		}
		if multi {
			comma, err = s.scanContinue()
			if err != nil {
				return nil, err
			}
			if comma {
				err = s.scanStart()
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		return cs, nil
	}
}
func (s *scanner) scanPolydata() ([][]Coord, error) {
	err := s.scanStart()
	if err != nil {
		return nil, err
	}
	var poly [][]Coord
	var cs []Coord
	var comma bool
	for {
		cs, err = s.scanCoords(false)
		if err != nil {
			return nil, err
		}
		if len(cs) < 4 {
			return nil, fmt.Errorf("a polygon ring must have at least 4 points, got %d", len(cs))
		}
		if cs[0] != cs[len(cs)-1] {
			return nil, fmt.Errorf("a polygon ring must be closed")
		}
		poly = append(poly, cs)
		comma, err = s.scanContinue()
		if err != nil {
			return nil, err
		}
		if comma {
			continue
		}
		return poly, nil
	}
}

func (s *scanner) scanMultiPolydata() ([][][]Coord, error) {
	err := s.scanStart()
	if err != nil {
		return nil, err
	}
	var multi [][][]Coord
	var poly [][]Coord
	var comma bool
	for {
		poly, err = s.scanPolydata()
		if err != nil {
			return nil, err
		}
		multi = append(multi, poly)
		comma, err = s.scanContinue()
		if err != nil {
			return nil, err
		}
		if comma {
			continue
		}
		return multi, nil
	}
}

func (s *scanner) scanGeom() (Geom, error) {
	ident, err := s.scanIdent()
	if err != nil {
		return nil, err
	}
	switch zmident, _ := s.scanIdent(); zmident {
	case "Z":
		s.opt = Z
	case "M":
		s.opt = M
	case "ZM":
		s.opt = ZM
	default:
		s.opt = 0
	}
	var g Geom
	switch ident {
	case "POINT", "MULTIPOINT", "LINESTRING":
		var cs []Coord
		cs, err = s.scanCoords(ident == "MULTIPOINT")
		if err != nil {
			break
		}
		switch ident {
		case "POINT":
			if len(cs) != 1 {
				return nil, fmt.Errorf("expected 1 got %d points", len(cs))
			}
			g = &Point{cs[0], s.opt}
		case "MULTIPOINT":
			g = &MultiPoint{cs, s.opt}
		case "LINESTRING":
			g = &LineString{cs, s.opt}
		}
	case "POLYGON":
		var rings [][]Coord
		rings, err = s.scanPolydata()
		if err != nil {
			break
		}
		g = &Polygon{rings, s.opt}
	case "MULTIPOLYGON":
		var multi [][][]Coord
		multi, err = s.scanMultiPolydata()
		if err != nil {
			break
		}
		g = &MultiPolygon{multi, s.opt}
	default:
		err = fmt.Errorf("unknown geom '%s'", ident)
	}
	if err != nil {
		return nil, err
	}
	return g, nil
}
