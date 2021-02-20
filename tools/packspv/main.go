package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var shaderFile string
var packageName string
var manifestFile string

func main() {
	flag.StringVar(&packageName, "p", "main", "Package name")
	flag.StringVar(&shaderFile, "f", "shaders.go", "Go file name")
	flag.StringVar(&manifestFile, "m", "", "Manifest file name")
	flag.Parse()
	fmt.Println("Packspv is depreciated tool. Use go:embed to import shaders and other attachments to go modules!")
	if flag.NArg() < 1 {
		usage()
	}
	bf := &bytes.Buffer{}
	writeHeader(bf)
	var err error
	if len(manifestFile) > 0 {
		err = readManifest(bf)
		if err != nil {
			log.Fatal("Error reading manifest: ", err)
		}
	} else {
		files, err := ioutil.ReadDir(flag.Arg(0))
		if err != nil {
			log.Fatal("Error reading directory: ", err)
		}
		for _, f := range files {
			err = writeDefaultFile(bf, f)
			if err != nil {
				log.Fatal("Error writing file: ", f.Name(), ": ", err)
			}
		}
	}
	shName := filepath.Join(flag.Arg(0), shaderFile)
	err = ioutil.WriteFile(shName, bf.Bytes(), 0660)
	if err != nil {
		log.Fatal("Error writing file: ", shName, ": ", err)
	}
	fmt.Println("Generated ", shName)
}

func readManifest(bf *bytes.Buffer) error {
	fManifest, err := os.Open(manifestFile)
	if err != nil {
		return err
	}
	defer fManifest.Close()
	sc := bufio.NewScanner(fManifest)
	for sc.Scan() {
		t := sc.Text()
		idx := strings.IndexRune(t, '#')
		if idx >= 0 {
			t = t[:idx]
		}
		t = strings.Trim(t, " \t")
		err = writeFile(bf, t)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeDefaultFile(bf *bytes.Buffer, info os.FileInfo) error {
	if !strings.HasSuffix(strings.ToLower(info.Name()), ".spv") {
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".bin") {
			return nil
		}
	}
	return writeFile(bf, info.Name())
}

func writeFile(bf *bytes.Buffer, fName string) error {
	varName := quoteFileName(fName)
	bytes, err := ioutil.ReadFile(filepath.Join(flag.Arg(0), fName))
	if err != nil {
		return err
	}
	_, _ = bf.WriteString("var ")
	_, _ = bf.WriteString(varName)
	_, _ = bf.WriteString("= []byte {")
	for idx, b := range bytes {
		if idx%40 == 39 {
			bf.WriteString("\n    ")
		} else {
			bf.WriteRune(' ')
		}
		bf.WriteString(strconv.Itoa(int(b)))
		bf.WriteString(",")
	}
	_, _ = bf.WriteString("\n}\n\n")
	return nil
}

func quoteFileName(fName string) string {
	return strings.Replace(strings.ToLower(filepath.Base(fName)), ".", "_", -1)
}

func writeHeader(bf *bytes.Buffer) {
	_, _ = bf.WriteString("//\n\n\n// Autogenerated file. Do not edit!\n package ")
	_, _ = bf.WriteString(packageName)
	_, _ = bf.WriteString("\n\n")
}

func usage() {
	fmt.Println("Usage: packsvp directory")
	fmt.Println("  ver 1.0.1")
	os.Exit(1)
}
