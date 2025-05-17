// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geometry

import (
	"fmt"
	parser2 "github.com/beetlebugorg/go-dims/internal/geometry/parser"
	"math"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/davidbyttow/govips/v2/vips"
)

type Flags struct {
	WidthPercent   bool
	HeightPercent  bool
	OffsetXPercent bool
	OffsetYPercent bool
	Force          bool
	OnlyGrow       bool
	OnlyShrink     bool
	Fill           bool
}

type Geometry struct {
	Width  float64
	Height float64
	X      int
	Y      int
	Flags  Flags
}

// ParseGeometry parse a geometry string in the form of "WIDTHxHEIGHT{+}X{+}Y{!<>}"
//
// WIDTH and HEIGHT are integers, and can be followed by '%' to indicate percentage.
//
// One WIDTH or HEIGHT is required.
//
// X and Y are offsets, and must be preceded by '+'.
//
// The '!' flag forces the image to be resized to the specified dimensions.
//
// The '<' flag only allows the image to be resized if it is smaller than the specified dimensions.
//
// The '>' flag only allows the image to be resized if it is larger than the specified dimensions.
//
// The '^' flag forces the image to fill the smallest dimension of the specified dimensions.
//
// Examples:
//
// "100x200" - width 100, height 200
// "50%x50%" - width 50%, height 50%
// "300x" - width 300
// "x400" - height 400
// "100x200+50+50%" - width 100, height 200, x offset 50, y offset 50%
// "+50+50" - x offset 50, y offset 50, width and height are 100% of the http of the image
// "100x100%+50+50" - width 100, height 100%, x offset 50, y offset 50
func ParseGeometry(geometry string) (Geometry, error) {
	is := antlr.NewInputStream(geometry)

	var errorListener = errorListener{}

	lexer := parser2.NewGeometryLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&errorListener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	var p = parser2.NewGeometryParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(&errorListener)
	var g = &geometryListener{
		Geometry: &Geometry{},
	}

	antlr.ParseTreeWalkerDefault.Walk(g, p.Start_())

	if len(errorListener.Errors) > 0 {
		return Geometry{}, errorListener.Errors[0]
	}

	return *g.Geometry, nil
}

// ApplyMeta returns a geometry that is modified as determined
// by the meta characters:  %, !, <, >, ^ in relation to the provided image.
//
// Final image dimensions are adjusted so as to preserve the aspect ratio as
// much as possible, while generating a integer (pixel) size, and fitting the
// image within the specified geometry width and height.
//
// Flags are interpreted...
//
//	%   geometry size is given percentage of original width and height given
//	!   do not try to preserve aspect ratio
//	<   only enlarge images smaller that geometry
//	>   only shrink images larger than geometry
//	^   fill given area
//
// A description of each parameter follows:
//
//	o geometry:  The geometry string (e.g. "100x100+10+10").
//	o x,y:  The x and y offset, set according to the geometry specification.
//	o width,height:  The width and height of original image, modified by
//	  the given geometry specification.
func (g Geometry) ApplyMeta(image *vips.ImageRef) Geometry {
	// Copy the original geometry
	var meta = &g

	// Get original image dimensions
	origWidth := float64(image.Width())
	origHeight := float64(image.Height())
	requestedWidth := meta.Width
	requestedHeight := meta.Height

	// Set width and height to original image dimensions if not specified
	if meta.Width == 0 {
		meta.Width = origWidth
	}

	if meta.Height == 0 {
		meta.Height = origHeight
	}

	// Apply width and height percentage if specified
	if g.Flags.WidthPercent {
		meta.Width = origWidth * g.Width / 100.0
		meta.Flags.WidthPercent = false
	}
	if g.Flags.HeightPercent {
		meta.Height = origHeight * g.Height / 100.0
		meta.Flags.HeightPercent = false
	}

	// Apply offset x and y percentage if specified
	if g.Flags.OffsetXPercent {
		meta.X = int(origWidth * float64(g.X) / 100.0)
		meta.Flags.OffsetXPercent = false
	}

	if g.Flags.OffsetYPercent {
		meta.Y = int(origHeight * float64(g.Y) / 100.0)
		meta.Flags.OffsetYPercent = false
	}

	if g.Flags.Fill {
		scaleX := meta.Width / origWidth
		scaleY := meta.Height / origHeight
		scale := math.Max(scaleX, scaleY)

		scaledWidth := origWidth * scale
		scaledHeight := origHeight * scale

		meta.Width = scaledWidth
		meta.Height = scaledHeight
	}

	// Apply aspect ratio if not forced
	if !g.Flags.Force {
		if requestedWidth != 0 || requestedHeight != 0 {
			// Fill the width and height from the original image ratio
			ratio := origWidth / origHeight
			if meta.Width/meta.Height > ratio {
				meta.Width = meta.Height * ratio
			} else {
				meta.Height = meta.Width / ratio
			}
		}
	}

	// Apply enlarge smaller images flag
	if g.Flags.OnlyGrow && (origWidth < meta.Width || origHeight < meta.Height) {
		if origWidth < meta.Width {
			meta.Width = origWidth
		}
		if origHeight < meta.Height {
			meta.Height = origHeight
		}
	}

	// Apply shrink larger images flag
	if g.Flags.OnlyShrink && (origWidth > meta.Width || origHeight > meta.Height) {
		if origWidth > meta.Width {
			meta.Width = origWidth
		}
		if origHeight > meta.Height {
			meta.Height = origHeight
		}
	}

	return *meta
}

func (g Geometry) String() string {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(int(g.Width)))
	if g.Flags.WidthPercent {
		builder.WriteString("%")
	}

	if g.Height > 0 {
		builder.WriteString("x")
		builder.WriteString(strconv.Itoa(int(g.Height)))

		if g.Flags.HeightPercent {
			builder.WriteString("%")
		}
	}

	if g.X >= 0 {
		builder.WriteString("+")
		builder.WriteString(strconv.Itoa(g.X))

		if g.Flags.OffsetXPercent {
			builder.WriteString("%")
		}
	}

	if g.Y >= 0 {
		builder.WriteString("+")
		builder.WriteString(strconv.Itoa(g.Y))

		if g.Flags.OffsetYPercent {
			builder.WriteString("%")
		}
	}

	return builder.String()
}

//-- ErrorListener

type syntaxError struct {
	line   int
	column int
	msg    string
}

func (e syntaxError) Error() string {
	return fmt.Sprintf("syntax error at column %d: %s", e.column, e.msg)
}

type errorListener struct {
	*antlr.DefaultErrorListener
	Errors []syntaxError
}

func (g *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	g.Errors = append(g.Errors, syntaxError{line, column, msg})
}

//-- GeometryListener

type geometryListener struct {
	*parser2.BaseGeometryListener
	*Geometry
}

func (g *geometryListener) ExitWidth(c *parser2.WidthContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Width, _ = strconv.ParseFloat(c.NUMBER().GetText(), 64)

	if c.PERCENT() != nil {
		g.Flags.WidthPercent = true
	}
}

func (g *geometryListener) ExitHeight(c *parser2.HeightContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Height, _ = strconv.ParseFloat(c.NUMBER().GetText(), 64)

	if c.PERCENT() != nil {
		g.Flags.HeightPercent = true
	}
}

func (g *geometryListener) ExitOffsetx(c *parser2.OffsetxContext) {
	if c.NUMBER() == nil {
		return
	}

	g.X, _ = strconv.Atoi(c.NUMBER().GetText())

	if c.PERCENT() != nil {
		g.Flags.OffsetXPercent = true
	}
}

func (g *geometryListener) ExitOffsety(c *parser2.OffsetyContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Y, _ = strconv.Atoi(c.NUMBER().GetText())

	if c.PERCENT() != nil {
		g.Flags.OffsetYPercent = true
	}
}

func (g *geometryListener) ExitFlags(c *parser2.FlagsContext) {
	if c.BANG() != nil {
		g.Flags.Force = true
	}

	if c.GT() != nil {
		g.Flags.OnlyGrow = true
	}

	if c.LT() != nil {
		g.Flags.OnlyShrink = true
	}

	if c.HAT() != nil {
		g.Flags.Fill = true
	}
}
