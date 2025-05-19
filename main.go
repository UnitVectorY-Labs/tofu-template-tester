package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	flagList        = flag.Bool("list-params", false, "List all template variables")
	flagIn          = flag.String("in", "", "Path to input template file (default: STDIN)")
	flagProps       = flag.String("properties", "", "Path to properties file (key=value)")
	flagInteractive = flag.Bool("interactive", false, "Prompt for missing variables interactively")
	flagOut         = flag.String("out", "", "Path to write output (default: STDOUT)")
)

// regex to find Terraform template placeholders like ${NAME}
var varRe = regexp.MustCompile(`\${\s*([A-Za-z0-9_]+)\s*}`)

func main() {
	flag.Parse()

	input, err := readInput(*flagIn)
	if err != nil {
		exitErr(err)
	}

	// List parameters
	if *flagList {
		names := listParams(input)
		for _, n := range names {
			fmt.Println(n)
		}
		return
	}

	// Load properties
	props := make(map[string]string)
	if *flagProps != "" {
		props, err = loadProperties(*flagProps)
		if err != nil {
			exitErr(err)
		}
	}

	// Interactive prompts
	if *flagInteractive {
		props = promptForVars(listParams(input), props)
	}

	// Process template
	output, err := processTemplate(input, props)
	if err != nil {
		exitErr(err)
	}

	// Write output
	if err := writeOutput(output, *flagOut); err != nil {
		exitErr(err)
	}
}

func readInput(path string) (string, error) {
	var reader io.Reader
	if path == "" {
		reader = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer f.Close()
		reader = f
	}
	b, err := io.ReadAll(reader)
	return string(b), err
}

func listParams(template string) []string {
	matches := varRe.FindAllStringSubmatch(template, -1)
	set := make(map[string]struct{})
	for _, m := range matches {
		set[m[1]] = struct{}{}
	}
	vars := make([]string, 0, len(set))
	for k := range set {
		vars = append(vars, k)
	}
	sort.Strings(vars)
	return vars
}

func loadProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	props := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid property line: " + line)
		}
		props[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return props, scanner.Err()
}

func promptForVars(vars []string, existing map[string]string) map[string]string {
	s := bufio.NewScanner(os.Stdin)
	for _, v := range vars {
		if _, ok := existing[v]; !ok {
			fmt.Printf("%s: ", v)
			if !s.Scan() {
				exitErr(errors.New("input aborted"))
			}
			existing[v] = s.Text()
		}
	}
	return existing
}

func processTemplate(tmpl string, props map[string]string) (string, error) {
	return varRe.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := varRe.FindStringSubmatch(match)[1]
		val, ok := props[key]
		if !ok {
			exitErr(fmt.Errorf("missing property: %s", key))
		}
		return val
	}), nil
}

func writeOutput(output, path string) error {
	var writer io.Writer
	if path == "" {
		writer = os.Stdout
	} else {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		writer = f
	}
	_, err := writer.Write([]byte(output))
	return err
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
