package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var (
	path    = flag.String("path", "/", "find path")
	pattern = flag.String("pattern", ".*go$", "grep pattern")
	workers = runtime.NumCPU()
)

func parseArgs() (string, string) {
	flag.Parse()
	return *path, *pattern
}

func find(path string) <-chan string {
	out := make(chan string, 1000)
	go func() {
		filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
			out <- file
			return nil
		})
		close(out)
	}()
	return out
}

func grep(pattern string, in <-chan string) <-chan string {
	out := make(chan string, cap(in))

	go func() {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}
		for file := range in {
			if reg.MatchString(file) {
				out <- file
			}
		}
		close(out)
	}()
	return out
}

func display(in <-chan string) <-chan struct{} {
	out := make(chan struct{})
	go func() {
		for file := range in {
			c, err := lineCount(file)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%8d %s\n", c, file)
		}
		out <- struct{}{}
	}()
	return out
}

func lineCount(file string) (int, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}

	if info.Mode().IsDir() {
		return 0, fmt.Errorf("%s is a directory", file)
	}

	count := 0
	buf := make([]byte, 1024*8)
	newLine := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		count += bytes.Count(buf[:c], newLine)

		if err == io.EOF {
			break
		}
	}
	return count, nil
}

func main() {
	runtime.GOMAXPROCS(workers)
	path, pattern := parseArgs()
	<-display(grep(pattern, find(path)))
}
