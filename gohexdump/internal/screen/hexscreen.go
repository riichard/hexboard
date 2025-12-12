
package screen

import (
	"post6.net/gohexdump/internal/font"
)

type HexScreen interface {
	TextScreen

	WriteRawTitle(g []font.Glyph, start int)
	WriteRawHexField(g []font.Glyph, field int)
	WriteRawAsciiField(g font.Glyph, field int)
	WriteRawOffset(g []font.Glyph, line int)


	WriteTitle(s string, start int)
	WriteHexField(s string, field int)
	WriteAsciiField(s string, field int)
	WriteOffset(s string, line int)

}

type hexScreen struct {
	textScreen

}

var hexConfig = Configuration{
	{ 0, 0, HorizontalPanel },
	{ 0, 1, HorizontalPanel },
	{ 0, 2, HorizontalPanel },
	{ 0, 3, HorizontalPanel },
}

const hexStartRow = 0
var hexColumns = []int{13,16,19,22,25,28,31,34,41,44,47,50,53,56,59,62}
const offsetColumn = 0
const asciiColumn = 69


func NewHexScreen() HexScreen {

	s := new(hexScreen)
	s.init(hexConfig)
	return s
}


func (t *hexScreen) WriteRawTitle(g []font.Glyph, start int) {
	if start < 0 || start >= 64 {
		return
	}
	l := 64-start
	if l > len(g) {
		l = len(g)
	}
	t.WriteRawAt(g[:l], start, 0)
}

func (t *hexScreen) WriteRawHexField(g []font.Glyph, field int) {
	twodigits := make([]font.Glyph, 2)
	copy(twodigits, g)
	t.WriteRawAt(twodigits, hexColumns[field%16], hexStartRow+(field/16))
}

func (t *hexScreen) WriteRawAsciiField(g font.Glyph, field int) {
	g_a := []font.Glyph{ g }
	if 0 <= field && field < 256 {
		t.WriteRawAt(g_a,  asciiColumn+(field%16), hexStartRow+(field/16))
	}
}

func (t *hexScreen) WriteRawOffset(g []font.Glyph, line int) {
	offsetdigits := make([]font.Glyph, 8)
	copy(offsetdigits, g)
	if 0 <= line && line < 16 {
		t.WriteRawAt(offsetdigits, offsetColumn, line+hexStartRow)
	}
}


func (t *hexScreen) WriteTitle(s string, start int) {
	t.WriteRawTitle(t.font.Glyphs(s), start)
}

func (t *hexScreen) WriteHexField(s string, field int) {

	t.WriteRawHexField(t.font.Glyphs(s), field)
}

func (t *hexScreen) WriteAsciiField(s string, field int) {
	t.WriteRawAsciiField(t.font.Glyphs(s)[0], field)
}

func (t *hexScreen) WriteOffset(s string, line int) {
	t.WriteRawOffset(t.font.Glyphs(s), line)
}


