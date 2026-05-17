package text

import (
	"bufio"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed fonts/*.flf
var fontFiles embed.FS

type figFont struct {
	name      string
	height    int
	hardblank rune
	glyphs    map[rune][]string
}

var fontCache = map[string]*figFont{}

func Render(input string, styleName string, width int) (string, error) {
	font, err := loadFont(styleName)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("text cannot be empty")
	}

	lines := make([]string, font.height)
	for _, r := range input {
		glyph, ok := font.glyphs[r]
		if !ok {
			glyph = font.glyphs['?']
		}
		for row := 0; row < font.height; row++ {
			lines[row] += glyph[row]
		}
	}

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n", nil
}

func StyleNames() []string {
	entries, _ := fontFiles.ReadDir("fonts")
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".flf") {
			continue
		}
		names = append(names, slug(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))))
	}
	sort.Strings(names)
	return names
}

func loadFont(styleName string) (*figFont, error) {
	key := slug(styleName)
	if font := fontCache[key]; font != nil {
		return font, nil
	}

	fileName, err := fontFileName(key)
	if err != nil {
		return nil, err
	}

	data, err := fontFiles.ReadFile("fonts/" + fileName)
	if err != nil {
		return nil, err
	}
	font, err := parseFLF(key, string(data))
	if err != nil {
		return nil, err
	}
	fontCache[key] = font
	return font, nil
}

func fontFileName(key string) (string, error) {
	entries, _ := fontFiles.ReadDir("fonts")
	for _, entry := range entries {
		base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if slug(base) == key {
			return entry.Name(), nil
		}
	}
	return "", fmt.Errorf("unknown style %q; run `gli text --list`", key)
}

func parseFLF(name string, content string) (*figFont, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	if !scanner.Scan() {
		return nil, fmt.Errorf("empty style %q", name)
	}
	header := strings.Fields(scanner.Text())
	if len(header) < 6 || !strings.HasPrefix(header[0], "flf2a") {
		return nil, fmt.Errorf("invalid style %q", name)
	}

	signature := []rune(header[0])
	hardblank := signature[len(signature)-1]
	height, err := strconv.Atoi(header[1])
	if err != nil {
		return nil, fmt.Errorf("invalid height in style %q", name)
	}
	commentLines, err := strconv.Atoi(header[5])
	if err != nil {
		return nil, fmt.Errorf("invalid comments in style %q", name)
	}

	for i := 0; i < commentLines && scanner.Scan(); i++ {
	}

	font := &figFont{
		name:      name,
		height:    height,
		hardblank: hardblank,
		glyphs:    make(map[rune][]string, 95),
	}

	for ch := rune(32); ch <= 126; ch++ {
		glyph := make([]string, height)
		var endmark rune
		for row := 0; row < height; row++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("style %q ended while reading glyph %q", name, ch)
			}
			line := strings.TrimRight(scanner.Text(), "\r")
			if endmark == 0 {
				endmark = lastRune(line)
			}
			line = trimEndmark(line, endmark)
			line = strings.ReplaceAll(line, string(hardblank), " ")
			glyph[row] = line
		}
		font.glyphs[ch] = glyph
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return font, nil
}

func lastRune(value string) rune {
	var last rune
	for _, r := range value {
		last = r
	}
	return last
}

func trimEndmark(value string, endmark rune) string {
	for strings.HasSuffix(value, string(endmark)) {
		value = strings.TrimSuffix(value, string(endmark))
	}
	return value
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return value
}
