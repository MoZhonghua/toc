package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func splitTitle(line string) (int, string) {
	count := 0
	title := ""
	for i, c := range line {
		if c != '#' {
			title = strings.TrimSpace(line[i:])
			break
		}
		count++
	}
	if count <= 1 {
		return 0, ""
	}

	return count - 1, title
}

type link struct {
	level int
	title string
}

const tocTitle = "Table of Contents"

func main() {
	var r io.Reader
	if len(os.Args) >= 2 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		r = bytes.NewReader(data)
	} else {
		r = os.Stdin
	}

	br := bufio.NewScanner(r)
	br.Buffer(nil, 1024*1024)
	lines := make([]string, 0)
	for br.Scan() {
		lines = append(lines, br.Text())
	}
	err := br.Err()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read file: %s\n", err.Error())
		os.Exit(1)
	}

	links := make([]link, 0)
	body := bytes.NewBuffer(nil)
	prelineIsAnchor := false
	insideOldToc := false
	for _, l := range lines {
		level, title := splitTitle(l)
		if level == 1 {
			if title == tocTitle {
				insideOldToc = true
			} else {
				insideOldToc = false
			}
		}

		if level == 0 {
			if !insideOldToc {
				fmt.Fprintln(body, l)
			}

			if strings.HasPrefix(l, "<a name") {
				prelineIsAnchor = true
			}
			continue
		}

		if !insideOldToc {
			links = append(links, link{level, title})
			if !prelineIsAnchor {
				fmt.Fprintf(body, "<a name=\"%s\"></a>\n", title)
			}
			fmt.Fprintln(body, l)
		}

		prelineIsAnchor = false
	}

	fmt.Printf("## %s\n\n", tocTitle)
	for _, l := range links {
		for i := 1; i < l.level; i++ {
			fmt.Print("  ")
		}
		fmt.Printf("* [%s](#%s)\n", l.title, l.title)
	}

	fmt.Printf("\n\n")
	io.Copy(os.Stdout, body)
}
