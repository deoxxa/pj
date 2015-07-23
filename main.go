package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	inputFile  = flag.String("input", "-", "Input file (default stdin).")
	outputFile = flag.String("output", "-", "Output file (default stdout).")
	width      = flag.Uint("width", 80, "Soft width limit for output.")
	indent     = flag.Uint("indent", 2, "Indent width.")
	sortKeys   = flag.Bool("sort_keys", false, "Sort object keys lexicographically.")
)

func main() {
	flag.Parse()

	input := os.Stdin
	if *inputFile != "-" {
		fd, err := os.OpenFile(*inputFile, os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}
		input = fd
	}
	defer input.Close()

	output := os.Stdout
	if *outputFile != "-" {
		fd, err := os.OpenFile(*outputFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		output = fd
	}
	defer output.Close()

	var v interface{}
	if err := json.NewDecoder(input).Decode(&v); err != nil {
		panic(err)
	}

	if err := encodeJSON(v, int(*width), output); err != nil {
		panic(err)
	}
}

func encodeJSON(v interface{}, width int, output io.Writer) error {
	d, err := formatIndent(v, width, 0, 0)
	if err != nil {
		return err
	}

	r := len(d)
	for r > 0 {
		n, err := output.Write([]byte(d))
		if err != nil {
			return err
		}

		r -= n
	}

	return nil
}

func formatIndent(v interface{}, width, level, additional int) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	switch v.(type) {
	case string, bool:
		return json.Marshal(v)
	case float64:
		return []byte(fmt.Sprintf("%.f", v)), nil
	case []interface{}, map[string]interface{}:
		if d, err := formatOneLine(v); err != nil {
			return nil, err
		} else if level*int(*indent)+additional+len(d) <= width {
			return d, nil
		}

		if v, ok := v.([]interface{}); ok {
			return formatArray(v, width, level)
		}

		if v, ok := v.(map[string]interface{}); ok {
			return formatObject(v, width, level)
		}
	}

	return nil, fmt.Errorf("couldn't encode type %T", v)
}

func formatArray(a []interface{}, width, level int) ([]byte, error) {
	bits := []string{"["}

	j := len(a)
	for i, v := range a {
		d, err := formatIndent(v, width, level+1, 0)
		if err != nil {
			return nil, err
		}

		suffix := ""
		if i < j-1 {
			suffix = ","
		}

		bits = append(bits, strings.Repeat(" ", (level+1)*int(*indent))+string(d)+suffix)
	}

	bits = append(bits, strings.Repeat(" ", level*int(*indent))+"]")

	return []byte(strings.Join(bits, "\n")), nil
}

func formatObject(m map[string]interface{}, width, level int) ([]byte, error) {
	bits := []string{"{"}

	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	if *sortKeys {
		sort.Strings(keys)
	}

	j := len(m)
	for i, k := range keys {
		v := m[k]

		kp, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}

		d, err := formatIndent(v, width, level+1, len(kp)+2)
		if err != nil {
			return nil, err
		}

		suffix := ""
		if i < j-1 {
			suffix = ","
		}

		bits = append(bits, strings.Repeat(" ", (level+1)*int(*indent))+string(kp)+": "+string(d)+suffix)

		i++
	}

	bits = append(bits, strings.Repeat(" ", level*int(*indent))+"}")

	return []byte(strings.Join(bits, "\n")), nil
}

func formatOneLine(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	switch v := v.(type) {
	case string, float64, bool:
		return json.Marshal(v)
	case []interface{}:
		return formatArrayOneLine(v)
	case map[string]interface{}:
		return formatObjectOneLine(v)
	}

	return nil, fmt.Errorf("can't format type %T", v)
}

func formatArrayOneLine(a []interface{}) ([]byte, error) {
	bits := []string{}

	for _, v := range a {
		d, err := formatOneLine(v)
		if err != nil {
			return nil, err
		}

		bits = append(bits, string(d))
	}

	return []byte("[" + strings.Join(bits, ", ") + "]"), nil
}

func formatObjectOneLine(m map[string]interface{}) ([]byte, error) {
	bits := []string{}

	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	if *sortKeys {
		sort.Strings(keys)
	}

	for _, k := range keys {
		v := m[k]

		kp, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}

		d, err := formatOneLine(v)
		if err != nil {
			return nil, err
		}

		bits = append(bits, string(kp)+": "+string(d))
	}

	return []byte("{" + strings.Join(bits, ", ") + "}"), nil
}
