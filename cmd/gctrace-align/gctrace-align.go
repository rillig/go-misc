// gctrace-align reads a Go GC log from the given file and aligns the columns.
//
// To produce the GC log, set GODEBUG=gctrace=1.
//
// See https://github.com/chewiebug/GCViewer/pull/196
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <gclogfile>", os.Args[0])
	}

	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	lines := parse(strings.Split(string(input), "\n"))
	table := formatTable(lines)
	for _, row := range table {
		println(row)
	}
}

type Line struct {
	relevant bool
	fields   []interface{}
}

func parse(in []string) []Line {
	var lines []Line
	for _, line := range in {
		if !strings.HasPrefix(line, "gc ") || !strings.HasSuffix(line, " P") {
			lines = append(lines, Line{false, []interface{}{line}})
			continue
		}

		matches := regexp.MustCompile(`(.*?)(\d+(?:\.\d+)?)`).FindAllStringSubmatchIndex(line, -1)
		if matches == nil {
			lines = append(lines, Line{false, []interface{}{line}})
			continue
		}

		var cells []interface{}
		for _, match := range matches {
			for i := 2; i < len(match); i += 2 {
				field := line[match[i]:match[i+1]]
				floatVal, err := strconv.ParseFloat(field, 64)

				var val interface{} = field
				switch {
				case err != nil:
					val = field
				case floatVal == float64(int(floatVal)):
					val = int(floatVal)
				default:
					val = floatVal
				}

				cells = append(cells, val)
			}
		}

		if lastMatch := matches[len(matches)-1]; lastMatch[1] != len(line) {
			cells = append(cells, line[lastMatch[1]:])
		}

		lines = append(lines, Line{true, cells})
	}
	return lines
}

func formatTable(table []Line) []string {
	cols := 0
	for _, row := range table {
		cols = intMax(cols, len(row.fields))
	}

	maxWidth := make([]int, cols, cols)  // Total width of the field
	maxBefore := make([]int, cols, cols) // Number of characters before the point
	maxAfter := make([]int, cols, cols)  // Number of characters after the point

	for _, line := range table {
		if line.relevant {
			fields := line.fields
			for i, field := range fields {
				str := fmt.Sprintf("%v", field)
				switch field.(type) {
				case int:
					maxBefore[i] = intMax(maxBefore[i], len(str))
				case float64:
					point := strings.IndexByte(str, '.')
					maxBefore[i] = intMax(maxBefore[i], point)
					maxAfter[i] = intMax(maxAfter[i], len(str)-(point+1))
				case string:
					maxWidth[i] = intMax(maxWidth[i], len(str))
				}
			}
		}
	}

	for i := 0; i < cols; i++ {
		dotWidth := 0
		if maxAfter[i] != 0 {
			dotWidth = 1
		}
		maxWidth[i] = intMax(maxWidth[i], maxBefore[i]+dotWidth+maxAfter[i])
	}

	var result []string
	for rowno, row := range table {
		_ = rowno
		if !row.relevant {
			result = append(result, row.fields[0].(string))
			continue
		}

		var resultRow []string
		fields := row.fields
		for i, obj := range fields {
			str := fmt.Sprintf("%v", obj)

			leftPad := 0
			switch obj.(type) {
			case int:
				leftPad = maxBefore[i] - len(str)
			case float64:
				leftPad = maxBefore[i] - strings.IndexByte(str, '.')
			}

			rightPad := 0
			if i != len(fields)-1 {
				rightPad = maxWidth[i] - leftPad - len(str)
			}

			padded := strings.Repeat(" ", leftPad) + str + strings.Repeat(" ", rightPad)
			resultRow = append(resultRow, padded)
		}
		result = append(result, strings.Join(resultRow, ""))
	}
	return result
}

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
