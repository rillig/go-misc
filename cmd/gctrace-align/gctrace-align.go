// gctrace-align reads a Go GC log from the given file and aligns the columns.
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

var gcLine = regexp.MustCompile("" +
	"(gc )(\\d+)" +
	"( @)(\\d+\\.\\d+)" +
	"(s )(\\d+)" +
	"(%: )(\\d+(?:\\.\\d+)?)" +
	"(\\+)(\\d+(?:\\.\\d+)?)" +
	"(\\+)(\\d+(?:\\.\\d+)?)" +
	"( ms clock, )(\\d+(?:\\.\\d+)?)" +
	"(\\+)(\\d+(?:\\.\\d+)?)" +
	"(/)(\\d+(?:\\.\\d+)?)" +
	"(/)(\\d+(?:\\.\\d+)?)" +
	"(\\+)(\\d+(?:\\.\\d+)?)" +
	"( ms cpu, )(\\d+)" +
	"(->)(\\d+)" +
	"(->)(\\d+)" +
	"( MB, )(\\d+)" +
	"( MB goal, )(\\d+)" +
	"( P)")

func main() {
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
	gc     bool
	fields []interface{}
}

func parse(in []string) []Line {
	var lines []Line
	for _, line := range in {
		match := gcLine.FindStringSubmatch(line)
		if match != nil {
			var cells []interface{}
			for _, field := range match[1:] {
				var val interface{}
				var err error
				val, err = strconv.Atoi(field)
				if err != nil {
					val, err = strconv.ParseFloat(field, 64)
					if err != nil {
						val = field
					}
				}
				cells = append(cells, val)
			}
			lines = append(lines, Line{true, cells})
		} else {
			lines = append(lines, Line{false, []interface{}{line}})
		}
	}
	return lines
}

func intMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatTable(table []Line) []string {
	cols := 0
	for _, row := range table {
		cols = intMax(cols, len(row.fields))
	}

	maxWidth := make([]int, cols, cols)
	maxBefore := make([]int, cols, cols)
	maxAfter := make([]int, cols, cols)

	for _, line := range table {
		if line.gc {
			fields := line.fields
			for i, field := range fields {
				str := fmt.Sprintf("%v", field)
				if _, ok := field.(int); ok {
					maxBefore[i] = intMax(maxBefore[i], len(str))
				} else if _, ok := field.(float64); ok {
					point := strings.IndexByte(str, '.')
					maxBefore[i] = intMax(maxBefore[i], point)
					maxAfter[i] = intMax(maxAfter[i], len(str)-(point+1))
				} else {
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
	for _, row := range table {
		if !row.gc {
			result = append(result, row.fields[0].(string))
			continue
		}

		var resultRow []string
		fields := row.fields
		for i, obj := range fields {
			str := fmt.Sprintf("%v", obj)
			switch obj.(type) {
			case int:
				str = strings.Repeat(" ", maxBefore[i]-len(str)) + str
			case float64:
				str = strings.Repeat(" ", maxBefore[i]-strings.IndexByte(str, '.')) + str
			}
			if i != len(fields)-1 {
				str += strings.Repeat(" ", maxWidth[i]-len(str))
			}

			resultRow = append(resultRow, str)
		}
		result = append(result, strings.Join(resultRow, ""))
	}
	return result
}
