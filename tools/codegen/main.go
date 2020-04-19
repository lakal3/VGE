package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	var outPath string
	flag.StringVar(&outPath, "out", "", "Output path")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}
	var buf *bytes.Buffer
	var out io.Writer
	out = os.Stdout
	if len(outPath) > 0 {
		buf = new(bytes.Buffer)
		out = buf
	}
	var err error
	switch flag.Arg(0) {
	case "gointerface":
		err = genGoInterface(out)
	case "gointerface2":
		err = genGoInterface2(out)
	case "goenums":
		err = genGoEnums(out)
	case "cppheader":
		err = genCppHeader(out)
	case "cppinterface":
		err = genCppInterface(out)
	}
	if err != nil {
		log.Fatal("Generate failed: ", err)
	}
	if len(outPath) > 0 {
		err = ioutil.WriteFile(outPath, buf.Bytes(), 0660)
		if err != nil {
			log.Fatal("Generate failed: ", err)
		} else {
			fmt.Println("Generated to ", outPath)
		}
	}

}

func usage() {
	fmt.Println("codegen option")
	fmt.Println("  gointerface - Generate Go interface from ")
	os.Exit(1)
}
