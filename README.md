wkt
===

Package wkt implements a simplified parser for Well Known Text.
It supports Point, MultiPoint, LineString, Polygon and MultiPolygon with both z and m coordinate parsing

	http://en.wikipedia.org/wiki/Well-known_text

Install
-------

	go get github.com/mb0/wkt

Basic Usage
-----------

You can find the online documentation at http://godoc.org/github.com/mb0/wkt

Example:

	geom, err := wkt.Parse([]byte(`POINT ZM(1.0 2.0 3.0 4.0)`))

License
-------
wkt is BSD licensed, Copyright (c) 2014 Martin Schnabel
