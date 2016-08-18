package main

import (
	"bufio"
	"bytes"
	"flag"
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

var idx = 0

var anchorPrefix = "toc_link"

func genUniqueId() string {
	idx++
	return fmt.Sprintf("%s_%03d", anchorPrefix, idx)
}

type link struct {
	level int
	title string
	id    string
}

const tocTitle = "Table of Contents"

var inplace bool
var legacy bool

func isAnchor(s string) bool {
	pattern := fmt.Sprintf("<a name=\"%s_", anchorPrefix)
	if strings.HasPrefix(s, pattern) {
		return true
	}

	if legacy && strings.HasPrefix(s, "<a name=") {
		return true
	}

	return false
}

func main() {
	flag.BoolVar(&inplace, "i", false, "inplace output")
	flag.BoolVar(&legacy, "l", false, "legacy anchor")
	flag.Parse()

	if inplace && len(flag.Args()) < 1 {
		fmt.Printf("no input file name to inplace write")
		return
	}

	var r io.Reader
	if len(flag.Args()) >= 1 {
		data, err := ioutil.ReadFile(flag.Args()[0])
		if err != nil {
			panic(err)
		}
		r = bytes.NewReader(data)
	} else {
		r = os.Stdin
	}

	br := bufio.NewScanner(r)
	// br.Buffer(nil, 1024*1024)
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

		if isAnchor(l) {
			continue
		}

		if level == 0 {
			if !insideOldToc {
				fmt.Fprintln(body, l)
			}

			continue
		}

		if !insideOldToc {
			id := genUniqueId()
			links = append(links, link{level, title, id})
			fmt.Fprintf(body, "<a name=\"%s\"></a>\n", id)
			fmt.Fprintln(body, l)
		}
	}

	toc := bytes.NewBuffer(nil)
	fmt.Fprintf(toc, "## %s\n\n", tocTitle)
	for _, l := range links {
		for i := 1; i < l.level; i++ {
			fmt.Fprint(toc, "  ")
		}
		fmt.Fprintf(toc, "* [%s](#%s)\n", l.title, l.id)
	}

	fmt.Fprintf(toc, "\n\n")

	if !inplace {
		io.Copy(os.Stdout, toc)
		io.Copy(os.Stdout, body)
		return
	}

	tmp, err := ioutil.TempFile(".", "XXXXXX.md")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(tmp, toc)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(tmp, body)
	if err != nil {
		panic(err)
	}

	tmpFileName := tmp.Name()
	tmp.Close()

	err = os.Rename(tmpFileName, flag.Args()[0])
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}
