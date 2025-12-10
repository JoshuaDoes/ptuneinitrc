package main

import (
	"strings"
)

type service struct {
	path string
	name string
	line string
	block []string
}

func newInitService(path, line string, block []string) *service {
	line = strings.TrimSpace(line)
	for i := 0; i < len(block); i++ {
		block[i] = strings.TrimSpace(block[i])
	}
	words := strings.Split(line, " ")
	return &service{
		path: path,
		name: words[1],
		line: line,
		block: block,
	}
}
