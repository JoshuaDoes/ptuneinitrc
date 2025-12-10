package main

import (
	"os/exec"
	"strings"
)

func getprop(key string) string {
	out, err := exec.Command("getprop", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func whitespacer(text string) string {
	data := []byte(text)
	for i := 0; i < len(data); i++ {
		if data[i] == '\t' {
			data[i] = ' '
		}
	}
	text = strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for i := 1; i < len(lines); i++ {
		for strings.HasPrefix(lines[i], "  ") {
			lines[i] = lines[i][1:]
		}
	}
	for strings.HasPrefix(lines[0], " ") {
		lines[0] = lines[0][1:]
	}
	return strings.Join(lines, "\n")
}

func nocomment(line string) string {
	if strings.Contains(line, "#") {
		parts := strings.SplitN(line, "#", 2)
		return strings.TrimSpace(parts[0])
	}
	return line
}