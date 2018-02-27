package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var hunkHeader = regexp.MustCompile(`^@@ -\d+,(\d+) \+\d+,(\d+) @@`)

// Tabdiff reads a unified diff from stdin and writes it to stdout,
// thereby changing the indentation from spaces to tabs.
func main() {
	if len(os.Args) > 1 {
		log.Fatal("usage: tabdiff (no arguments)")
	}

	text, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.SplitAfter(string(text), "\n")
	var del, add int
	for _, line := range lines {
		if del+add > 0 {
			if line == "" {
				line = " "
			}
			switch line[0] {
			case ' ', '\t':
				del--
				add--
				line = "\t" + line[1:]
			case '-':
				del--
				line = "-\t" + line[1:]
			case '+':
				add--
				line = "+\t" + line[1:]
			}
		} else if m := hunkHeader.FindStringSubmatch(line); m != nil {
			del, _ = strconv.Atoi(m[1])
			add, _ = strconv.Atoi(m[2])
		}
		io.WriteString(os.Stdout, line)
	}
}
