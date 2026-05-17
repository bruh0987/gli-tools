package text

import (
	"fmt"
	"strings"
	"unicode"
)

type style struct {
	Name      string
	Pixel     string
	Space     string
	Slant     bool
	Lower     bool
	Border    bool
	Tight     bool
	Underline bool
}

var styles = map[string]style{
	"standard": {Name: "standard", Pixel: "#", Space: " "},
	"slant":    {Name: "slant", Pixel: "/", Space: " ", Slant: true},
	"small":    {Name: "small", Pixel: "#", Space: " ", Tight: true},
	"big":      {Name: "big", Pixel: "##", Space: "  "},
	"block":    {Name: "block", Pixel: "#", Space: " "},
	"bubble":   {Name: "bubble", Pixel: "o", Space: " "},
	"digital":  {Name: "digital", Pixel: "8", Space: " "},
	"doom":     {Name: "doom", Pixel: "@", Space: " ", Border: true},
	"graffiti": {Name: "graffiti", Pixel: "$", Space: " ", Slant: true,
		Underline: true},
	"lean":     {Name: "lean", Pixel: "/", Space: " ", Slant: true, Tight: true},
	"mini":     {Name: "mini", Pixel: ".", Space: " ", Tight: true},
	"script":   {Name: "script", Pixel: "~", Space: " ", Slant: true},
	"shadow":   {Name: "shadow", Pixel: "#", Space: ".", Border: true},
	"smslant":  {Name: "smslant", Pixel: "/", Space: " ", Slant: true, Tight: true},
	"speed":    {Name: "speed", Pixel: ">", Space: " ", Slant: true},
	"starwars": {Name: "starwars", Pixel: "*", Space: " ", Border: true},
	"stop":     {Name: "stop", Pixel: "X", Space: " "},
	"straight": {Name: "straight", Pixel: "|", Space: " "},
	"term":     {Name: "term", Pixel: "#", Space: " ", Lower: true, Tight: true},
	"weird":    {Name: "weird", Pixel: "%", Space: "_", Slant: true},
}

func Render(input string, styleName string, width int) (string, error) {
	st, ok := styles[strings.ToLower(styleName)]
	if !ok {
		return "", fmt.Errorf("unknown style %q; run `gli text --list`", styleName)
	}
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("text cannot be empty")
	}

	lines := make([]string, 5)
	for _, r := range input {
		glyph := glyphFor(r)
		for row := 0; row < 5; row++ {
			part := strings.ReplaceAll(glyph[row], "#", st.Pixel)
			part = strings.ReplaceAll(part, " ", st.Space)
			if st.Lower {
				part = strings.ToLower(part)
			}
			gap := " "
			if st.Tight {
				gap = ""
			}
			lines[row] += part + gap
		}
	}

	if st.Slant {
		for i := range lines {
			lines[i] = strings.Repeat(" ", len(lines)-i-1) + lines[i]
		}
	}
	if st.Underline {
		maxLen := 0
		for _, line := range lines {
			maxLen = max(maxLen, len(line))
		}
		lines = append(lines, strings.Repeat("-", maxLen))
	}
	if st.Border {
		maxLen := 0
		for _, line := range lines {
			maxLen = max(maxLen, len(line))
		}
		border := "+" + strings.Repeat("-", maxLen+2) + "+"
		out := []string{border}
		for _, line := range lines {
			out = append(out, "| "+line+strings.Repeat(" ", maxLen-len(line))+" |")
		}
		out = append(out, border)
		lines = out
	}

	return strings.TrimRight(strings.Join(lines, "\n"), " \n") + "\n", nil
}

func glyphFor(r rune) [5]string {
	if r == ' ' {
		return [5]string{"   ", "   ", "   ", "   ", "   "}
	}
	r = unicode.ToUpper(r)
	if glyph, ok := glyphs[r]; ok {
		return glyph
	}
	return glyphs['?']
}

var glyphs = map[rune][5]string{
	'A': {" ### ", "#   #", "#####", "#   #", "#   #"},
	'B': {"#### ", "#   #", "#### ", "#   #", "#### "},
	'C': {" ####", "#    ", "#    ", "#    ", " ####"},
	'D': {"#### ", "#   #", "#   #", "#   #", "#### "},
	'E': {"#####", "#    ", "#### ", "#    ", "#####"},
	'F': {"#####", "#    ", "#### ", "#    ", "#    "},
	'G': {" ####", "#    ", "#  ##", "#   #", " ####"},
	'H': {"#   #", "#   #", "#####", "#   #", "#   #"},
	'I': {"#####", "  #  ", "  #  ", "  #  ", "#####"},
	'J': {"#####", "   # ", "   # ", "#  # ", " ##  "},
	'K': {"#   #", "#  # ", "###  ", "#  # ", "#   #"},
	'L': {"#    ", "#    ", "#    ", "#    ", "#####"},
	'M': {"#   #", "## ##", "# # #", "#   #", "#   #"},
	'N': {"#   #", "##  #", "# # #", "#  ##", "#   #"},
	'O': {" ### ", "#   #", "#   #", "#   #", " ### "},
	'P': {"#### ", "#   #", "#### ", "#    ", "#    "},
	'Q': {" ### ", "#   #", "# # #", "#  # ", " ## #"},
	'R': {"#### ", "#   #", "#### ", "#  # ", "#   #"},
	'S': {" ####", "#    ", " ### ", "    #", "#### "},
	'T': {"#####", "  #  ", "  #  ", "  #  ", "  #  "},
	'U': {"#   #", "#   #", "#   #", "#   #", " ### "},
	'V': {"#   #", "#   #", "#   #", " # # ", "  #  "},
	'W': {"#   #", "#   #", "# # #", "## ##", "#   #"},
	'X': {"#   #", " # # ", "  #  ", " # # ", "#   #"},
	'Y': {"#   #", " # # ", "  #  ", "  #  ", "  #  "},
	'Z': {"#####", "   # ", "  #  ", " #   ", "#####"},
	'0': {" ### ", "#  ##", "# # #", "##  #", " ### "},
	'1': {"  #  ", " ##  ", "  #  ", "  #  ", "#####"},
	'2': {" ### ", "#   #", "   # ", "  #  ", "#####"},
	'3': {"#### ", "    #", " ### ", "    #", "#### "},
	'4': {"#   #", "#   #", "#####", "    #", "    #"},
	'5': {"#####", "#    ", "#### ", "    #", "#### "},
	'6': {" ### ", "#    ", "#### ", "#   #", " ### "},
	'7': {"#####", "   # ", "  #  ", " #   ", "#    "},
	'8': {" ### ", "#   #", " ### ", "#   #", " ### "},
	'9': {" ### ", "#   #", " ####", "    #", " ### "},
	'!': {"  #  ", "  #  ", "  #  ", "     ", "  #  "},
	'?': {" ### ", "#   #", "   # ", "     ", "  #  "},
	'.': {"     ", "     ", "     ", "     ", "  #  "},
	',': {"     ", "     ", "     ", "  #  ", " #   "},
	'-': {"     ", "     ", "#####", "     ", "     "},
	'_': {"     ", "     ", "     ", "     ", "#####"},
	':': {"     ", "  #  ", "     ", "  #  ", "     "},
	'/': {"    #", "   # ", "  #  ", " #   ", "#    "},
}
