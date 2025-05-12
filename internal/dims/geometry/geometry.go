package geometry

import (
	"fmt"
	"math"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	"github.com/beetlebugorg/go-dims/internal/dims/geometry/parser"
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

// Parse a geometry string in the form of "WIDTHxHEIGHT{+}X{+}Y{!<>}"
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
// "+50+50" - x offset 50, y offset 50, width and height are 100% of the rest of the image
// "100x100%+50+50" - width 100, height 100%, x offset 50, y offset 50
func ParseGeometry(geometry string) (Geometry, error) {
	is := antlr.NewInputStream(geometry)

	var errorListener = errorListener{}

	lexer := parser.NewGeometryLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&errorListener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	var p = parser.NewGeometryParser(stream)
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
func (g *Geometry) ApplyMeta(image *vips.ImageRef) Geometry {
	// Copy the original geometry
	var meta = *g

	// Get original image dimensions
	origWidth := float64(image.Width())
	origHeight := float64(image.Height())
	requestedWidth := meta.Width
	requestedHeight := meta.Height

	// Set width and height to original image dimensions if not specified
	if meta.Width == 0 {
		meta.Width = float64(origWidth)
	}

	if meta.Height == 0 {
		meta.Height = float64(origHeight)
	}

	// Apply width and height percentage if specified
	if g.Flags.WidthPercent {
		meta.Width = float64(origWidth) * float64(g.Width) / 100.0
	}
	if g.Flags.HeightPercent {
		meta.Height = float64(origHeight) * float64(g.Height) / 100.0
	}

	// Apply offset x and y percentage if specified
	if g.Flags.OffsetXPercent {
		meta.X = int(float64(origWidth) * float64(g.X) / 100.0)
	}

	if g.Flags.OffsetYPercent {
		meta.Y = int(float64(origHeight) * float64(g.Y) / 100.0)
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
			ratio := float64(origWidth) / float64(origHeight)
			if float64(meta.Width)/float64(meta.Height) > ratio {
				meta.Width = float64(meta.Height) * ratio
			} else {
				meta.Height = float64(meta.Width) / ratio
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

	return meta
}

func (g Geometry) String() string {
	return fmt.Sprintf("%.0fx%0.f+%d+%d", g.Width, g.Height, g.X, g.Y)
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
	*parser.BaseGeometryListener
	*Geometry
}

func (g *geometryListener) ExitWidth(c *parser.WidthContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Width, _ = strconv.ParseFloat(c.NUMBER().GetText(), 64)

	if c.PERCENT() != nil {
		g.Flags.WidthPercent = true
	}
}

func (g *geometryListener) ExitHeight(c *parser.HeightContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Height, _ = strconv.ParseFloat(c.NUMBER().GetText(), 64)

	if c.PERCENT() != nil {
		g.Flags.HeightPercent = true
	}
}

func (g *geometryListener) ExitOffsetx(c *parser.OffsetxContext) {
	if c.NUMBER() == nil {
		return
	}

	g.X, _ = strconv.Atoi(c.NUMBER().GetText())

	if c.PERCENT() != nil {
		g.Flags.OffsetXPercent = true
	}
}

func (g *geometryListener) ExitOffsety(c *parser.OffsetyContext) {
	if c.NUMBER() == nil {
		return
	}

	g.Y, _ = strconv.Atoi(c.NUMBER().GetText())

	if c.PERCENT() != nil {
		g.Flags.OffsetYPercent = true
	}
}

func (g *geometryListener) ExitFlags(c *parser.FlagsContext) {
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
