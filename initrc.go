package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type initrc struct {
	name string
	path string
	file *os.File
}

func newInitRC(path string) (*initrc, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &initrc{
		name: filepath.Base(file.Name()),
		path: path,
		file: file,
	}, nil
}
func (irc *initrc) getServices() []*service {
	if irc.file == nil {
		return nil
	}
	data, err := io.ReadAll(irc.file)
	if err != nil {
		return nil
	}
	text := whitespacer(string(data))
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}
	services := make([]*service, 0)
	line := ""
	svcLines := make([]string, 0)
	for l := 0; l < len(lines); l++ {
		line = lines[l]
		if strings.HasPrefix(whitespacer(line), "#") {
			continue
		}
		//fmt.Printf("Parsing line %d: %s\n", l, line)
		if strings.HasPrefix(line, "service ") {
			//fmt.Printf("Found service at line %d: %s\n", l, line)
			svcLines = make([]string, 0)
			for i := l + 1; i < len(lines); i++ {
				l = i
				svcLine := lines[i]
				if strings.HasPrefix(whitespacer(svcLine), "#") {
					continue
				}
				if strings.HasPrefix(svcLine, " ") {
					//fmt.Printf("Parsing service line %d: %s\n", i, svcLine)
					svcLine = strings.TrimSpace(svcLine)
					words := strings.Split(svcLine, " ")
					switch words[0] {
					case "task_profiles", "ioprio", "priority", "writepid":
						svcLine = ""
					case "rlimit":
						if len(words) == 1 { //Malformed input??? Skip over the line for now
							svcLine = ""
						}
						if words[1] == "rtprio" || words[1] == "memlock" {
							svcLine = ""
						}
					}
					if svcLine == "" {
						//fmt.Printf("Ignoring service line %d\n", i)
						continue
					}
					svcLine = nocomment(svcLine)
					svcLines = append(svcLines, svcLine)
					continue
				}
				//fmt.Printf("Found end of service at line %d: %s\nHEX: %X\n", i, svcLine, svcLine)
				l--
				break
			}
			services = append(services, newInitService(irc.path, line, svcLines))
			svcLines = make([]string, 0)
		}
	}
	if len(svcLines) > 0 {
		//fmt.Println("Found end of service at end of file")
		services = append(services, newInitService(irc.path, line, svcLines))
	}
	return services
}
func (irc *initrc) close() {
	if irc.file != nil {
		irc.file.Close()
		irc.file = nil
	}
	irc.name = ""
	irc.path = ""
}

func findInitRCs() []*initrc {
	results := make([]string, 0)

	roots, err := os.ReadDir("/")
	if err == nil {
		for _, e := range roots {
			if e.Name() == "" {
				continue
			}
			path := filepath.Join("/", e.Name(), "etc/init")
			stat, err := os.Stat(path)
			if err == nil && stat.IsDir() {
				matches, _ := filepath.Glob(filepath.Join(path, "*.rc"))
				for _, m := range matches {
					results = append(results, m)
				}
				path = filepath.Join(path, "hw")
				stat, err = os.Stat(path)
				if err == nil && stat.IsDir() {
					matches, _ := filepath.Glob(filepath.Join(path, "*.rc"))
					for _, m := range matches {
						results = append(results, m)
					}
				}
			}
		}
	}

	roots, err = os.ReadDir("/apex")
	if err == nil {
		for _, e := range roots {
			if strings.Contains(e.Name(), "@") {
				continue //Skip version-mounted APEX package directories
			}
			path := filepath.Join("/apex", e.Name())
			if stat, err := os.Stat(path); err == nil && stat.IsDir() {
				_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if d.IsDir() {
						return nil
					}
					if strings.HasSuffix(d.Name(), ".rc") {
						results = append(results, p)
					}
					return nil
				})
			}
		}
	}

	sort.Strings(results)
	initrcs := make([]*initrc, 0)
	for _, r := range results {
		irc, err := newInitRC(r)
		if err == nil {
			initrcs = append(initrcs, irc)
		}
	}
	return initrcs
}