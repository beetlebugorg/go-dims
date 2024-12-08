package geometry

import (
	"fmt"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	"github.com/beetlebugorg/go-dims/internal/dims/geometry/parser"
)

type Flags struct {
	WidthPercent   bool
	HeightPercent  bool
	OffsetXPercent bool
	OffsetYPercent bool
	Force          bool
	OnlyGrow       bool
	OnlyShrink     bool
}

type Geometry struct {
	Width  int64
	Height int64
	X      int64
	Y      int64
	Flags  Flags
}

type geometryListener struct {
	*parser.BaseGeometryListener
	*Geometry
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
// Examples:
//
// "100x200" - width 100, height 200
// "50%x50%" - width 50%, height 50%
// "300x" - width 300
// "x400" - height 400
// "100x200+50+50%" - width 100, height 200, x offset 50, y offset 50%
// "+50+50" - x offset 50, y offset 50, width and height are 100% of the rest of the image
// "100x100%+50+50" - width 100, height 100%, x offset 50, y offset 50
func parseGeometry(geometry string) Geometry {
	stream := antlr.NewInputStream(geometry)
	lexer := parser.NewGeometryLexer(stream)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	var p = parser.NewGeometryParser(tokens)
	var g = &geometryListener{
		Geometry: &Geometry{},
	}

	antlr.ParseTreeWalkerDefault.Walk(g, p.Start_())

	return *g.Geometry
}

//-- GeometryListener

func (g *geometryListener) ExitWidth(c *parser.WidthContext) {
	var err error
	g.Width, err = strconv.ParseInt(c.NUMBER().GetText(), 10, 64)
	if err != nil {
		fmt.Println(err)
	}

	if c.Percent() != nil {
		g.Flags.WidthPercent = true
	}
}

func (g *geometryListener) ExitHeight(c *parser.HeightContext) {
	var err error
	g.Height, err = strconv.ParseInt(c.NUMBER().GetText(), 10, 64)
	if err != nil {
		fmt.Println(err)
	}

	if c.Percent() != nil {
		g.Flags.HeightPercent = true
	}
}

func (g *geometryListener) ExitOffsetx(c *parser.OffsetxContext) {
	var err error
	g.X, err = strconv.ParseInt(c.NUMBER().GetText(), 10, 64)
	if err != nil {
		fmt.Println(err)
	}

	if c.Percent() != nil {
		g.Flags.OffsetXPercent = true
	}
}

func (g *geometryListener) ExitOffsety(c *parser.OffsetyContext) {
	var err error
	g.Y, err = strconv.ParseInt(c.NUMBER().GetText(), 10, 64)
	if err != nil {
		fmt.Println(err)
	}

	if c.Percent() != nil {
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
}
