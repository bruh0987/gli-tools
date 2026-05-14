package lines

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type extensionStat struct {
	Extension string
	Files     int
	Lines     int64
}

type fileStat struct {
	Path  string
	Ext   string
	Lines int64
}

type config struct {
	root         string
	top          int
	useGitignore bool
	excludedExts map[string]struct{}
}

func Run(args []string, out io.Writer, errOut io.Writer) int {
	flags := flag.NewFlagSet("lines", flag.ContinueOnError)
	flags.SetOutput(errOut)

	top := flags.Int("top", 0, "display the top N files by line count")
	useGitignore := flags.Bool("gitignore", false, "ignore files matched by .gitignore")
	exclude := flags.String("exclude", "", "comma-separated extensions to exclude, like go,md,json")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	root := "."
	if flags.NArg() > 1 {
		fmt.Fprintln(errOut, "usage: gli lines [directory] [--top N] [--gitignore] [--exclude exts]")
		return 2
	}
	if flags.NArg() == 1 {
		root = flags.Arg(0)
	}

	cfg := config{
		root:         root,
		top:          *top,
		useGitignore: *useGitignore,
		excludedExts: parseExcludedExts(*exclude),
	}

	if cfg.top < 0 {
		fmt.Fprintln(errOut, "--top must be 0 or greater")
		return 2
	}

	result, err := countLines(cfg)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}

	printExtensionTable(out, result.extensions)
	if cfg.top > 0 {
		fmt.Fprintln(out)
		printTopFiles(out, result.files, cfg.top)
	}

	return 0
}

type result struct {
	extensions []extensionStat
	files      []fileStat
}

func countLines(cfg config) (result, error) {
	root, err := filepath.Abs(cfg.root)
	if err != nil {
		return result{}, err
	}

	info, err := os.Stat(root)
	if err != nil {
		return result{}, err
	}
	if !info.IsDir() {
		return result{}, fmt.Errorf("%s is not a directory", cfg.root)
	}

	matcher := gitignoreMatcher{}
	if cfg.useGitignore {
		matcher, err = loadGitignore(root)
		if err != nil {
			return result{}, err
		}
	}

	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}

	jobs := make(chan string, workers*4)
	results := make(chan fileStat, workers*4)
	errs := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for path := range jobs {
				stat, err := countFile(root, path)
				if err != nil {
					select {
					case errs <- err:
					default:
					}
					continue
				}
				results <- stat
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			if cfg.useGitignore && matcher.match(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}
		if cfg.useGitignore && matcher.match(rel, false) {
			return nil
		}
		if _, ok := cfg.excludedExts[extension(path)]; ok {
			return nil
		}

		jobs <- path
		return nil
	})
	close(jobs)

	byExt := make(map[string]*extensionStat)
	files := make([]fileStat, 0)
	for stat := range results {
		ext := stat.Ext
		current := byExt[ext]
		if current == nil {
			current = &extensionStat{Extension: ext}
			byExt[ext] = current
		}
		current.Files++
		current.Lines += stat.Lines
		files = append(files, stat)
	}

	if walkErr != nil {
		return result{}, walkErr
	}
	select {
	case err := <-errs:
		return result{}, err
	default:
	}

	extensions := make([]extensionStat, 0, len(byExt))
	for _, stat := range byExt {
		extensions = append(extensions, *stat)
	}

	sort.Slice(extensions, func(i, j int) bool {
		if extensions[i].Lines == extensions[j].Lines {
			return extensions[i].Extension < extensions[j].Extension
		}
		return extensions[i].Lines > extensions[j].Lines
	})

	sort.Slice(files, func(i, j int) bool {
		if files[i].Lines == files[j].Lines {
			return files[i].Path < files[j].Path
		}
		return files[i].Lines > files[j].Lines
	})

	return result{extensions: extensions, files: files}, nil
}

func countFile(root string, path string) (fileStat, error) {
	file, err := os.Open(path)
	if err != nil {
		return fileStat{}, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 1024*1024)
	var lines int64
	var sawAny bool
	var last byte
	buffer := make([]byte, 128*1024)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			sawAny = true
			chunk := buffer[:n]
			lines += int64(bytes.Count(chunk, []byte{'\n'}))
			last = chunk[n-1]
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fileStat{}, err
		}
	}

	if sawAny && last != '\n' {
		lines++
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fileStat{}, err
	}

	return fileStat{
		Path:  filepath.ToSlash(rel),
		Ext:   extension(path),
		Lines: lines,
	}, nil
}

func extension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "[none]"
	}
	return ext
}

func parseExcludedExts(raw string) map[string]struct{} {
	excluded := make(map[string]struct{})
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" {
			continue
		}
		if item == "none" || item == "[none]" {
			excluded["[none]"] = struct{}{}
			continue
		}
		if !strings.HasPrefix(item, ".") {
			item = "." + item
		}
		excluded[item] = struct{}{}
	}
	return excluded
}

func printExtensionTable(out io.Writer, stats []extensionStat) {
	fmt.Fprintln(out, "Extension  Files  Lines")
	fmt.Fprintln(out, "---------  -----  -----")
	for _, stat := range stats {
		fmt.Fprintf(out, "%-9s  %5d  %5d\n", stat.Extension, stat.Files, stat.Lines)
	}
}

func printTopFiles(out io.Writer, files []fileStat, top int) {
	if top > len(files) {
		top = len(files)
	}

	fmt.Fprintln(out, "Top Files")
	fmt.Fprintln(out, "---------")
	fmt.Fprintln(out, "Lines  File")
	fmt.Fprintln(out, "-----  ----")
	for i := 0; i < top; i++ {
		fmt.Fprintf(out, "%5d  %s\n", files[i].Lines, files[i].Path)
	}
}

type gitignoreMatcher struct {
	patterns []gitignorePattern
}

type gitignorePattern struct {
	raw      string
	pattern  string
	anchored bool
	dirOnly  bool
	hasSlash bool
}

func loadGitignore(root string) (gitignoreMatcher, error) {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if errors.Is(err, os.ErrNotExist) {
		return gitignoreMatcher{}, nil
	}
	if err != nil {
		return gitignoreMatcher{}, err
	}

	lines := strings.Split(string(data), "\n")
	patterns := make([]gitignorePattern, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		line = strings.TrimPrefix(line, `\`)
		pattern := filepath.ToSlash(line)
		dirOnly := strings.HasSuffix(pattern, "/")
		pattern = strings.Trim(pattern, "/")
		if pattern == "" {
			continue
		}

		patterns = append(patterns, gitignorePattern{
			raw:      line,
			pattern:  pattern,
			anchored: strings.HasPrefix(line, "/"),
			dirOnly:  dirOnly,
			hasSlash: strings.Contains(pattern, "/"),
		})
	}

	return gitignoreMatcher{patterns: patterns}, nil
}

func (m gitignoreMatcher) match(rel string, isDir bool) bool {
	rel = filepath.ToSlash(rel)
	for _, pattern := range m.patterns {
		if pattern.dirOnly && !isDir {
			continue
		}
		if pattern.match(rel) {
			return true
		}
	}
	return false
}

func (p gitignorePattern) match(rel string) bool {
	if p.anchored || p.hasSlash {
		return matchGlob(p.pattern, rel)
	}

	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if matchGlob(p.pattern, part) {
			return true
		}
	}
	return false
}

func matchGlob(pattern string, value string) bool {
	regex := globRegex(pattern)
	matched, err := filepath.Match(regex, value)
	if err == nil && matched {
		return true
	}
	return globSegments(pattern, value)
}

func globSegments(pattern string, value string) bool {
	patternParts := strings.Split(pattern, "/")
	valueParts := strings.Split(value, "/")
	return matchSegments(patternParts, valueParts)
}

func matchSegments(patterns []string, values []string) bool {
	if len(patterns) == 0 {
		return len(values) == 0
	}
	if patterns[0] == "**" {
		for i := 0; i <= len(values); i++ {
			if matchSegments(patterns[1:], values[i:]) {
				return true
			}
		}
		return false
	}
	if len(values) == 0 {
		return false
	}
	matched, err := filepath.Match(patterns[0], values[0])
	if err != nil || !matched {
		return false
	}
	return matchSegments(patterns[1:], values[1:])
}

func globRegex(pattern string) string {
	return strings.ReplaceAll(pattern, "**", "*")
}
