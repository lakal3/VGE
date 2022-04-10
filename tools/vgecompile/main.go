package main

// Tools to compile shader packs
// To install run: go install github.com/lakal3/vge/tools/vgecompile
//

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lakal3/vge/vge/shaders"
	"github.com/lakal3/vge/vge/vk"
	"log"
	"os"
	"path/filepath"
)

var debug bool
var device int

var config shaders.Config

func main() {
	flag.BoolVar(&debug, "debug", false, "Add debug layers")
	flag.IntVar(&device, "device", 0, "Device index")
	flag.Parse()
	if flag.NArg() < 2 {
		usage()
	}
	progJson, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal("Error reading program.json: ", err)
	}
	err = json.Unmarshal(progJson, &config)
	if err != nil {
		log.Fatal("Error unmarshalling program.json: ", err)
	}
	config.RootDir = filepath.Dir(flag.Arg(0))
	sp := &shaders.Pack{}
	err = genAll(sp)
	if err != nil {
		log.Fatal("Error compiling program.json: ", err)
	}
	fOut, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatal("Error writing ", flag.Arg(1), ": ", err)
	}
	defer fOut.Close()
	err = sp.Save(fOut)
	if err != nil {
		log.Println("Error writing ", flag.Arg(1), ": ", err)
	} else {
		fmt.Println("Compiled ", flag.Arg(0), " to ", flag.Arg(1))
	}
}

func genAll(sp *shaders.Pack) error {
	app, err := vk.NewApplication("VGEcompiler")
	if err != nil {
		return err
	}
	if debug {
		// Add validation layer if requested
		app.AddValidation()
	}
	app.Init()
	defer app.Dispose()
	pds := app.GetDevices()
	if len(pds) <= device {
		log.Fatal("No device ", device)
	}
	// Create a new device
	dev := app.NewDevice(int32(device))
	defer dev.Dispose()
	st := shaders.NewShaderTool()
	return st.CompileConfig(dev, sp, config)
}

func usage() {
	fmt.Println("vgecompile ", Version)
	fmt.Println("vgecompile programs.json output.bin")
	os.Exit(1)
}
