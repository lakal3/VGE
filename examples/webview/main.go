//
// Webview example uses Vulkan api to render 3D image into memory. Image is then packed to png and sent back to browser
// In browser html we have controls to change image rotation witch renders new image in server
//
// Webview example demonstrates on how to
// a) Use lower level of VGE to render without actual screen
//
// b) Manually render VGE scene

package main

import (
	"flag"
	"fmt"
	"github.com/lakal3/vge/vge/vk"
	"github.com/lakal3/vge/vge/vmodel"
	"image"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

var wwApp struct {
	owner   vk.Owner
	app     *vk.Application
	dev     *vk.Device
	pdIndex int
	model   *vmodel.Model
	rp      *vk.ForwardRenderPass
	debug   bool
}

// API Context used to initialize application.
type initContext struct {
}

func (i initContext) SetError(err error) {
	log.Fatal("Failed to initialize server: ", err)
}

func (i initContext) IsValid() bool {
	return true
}

func (i initContext) Begin(callName string) (atEnd func()) {
	return nil
}

func main() {
	flag.BoolVar(&wwApp.debug, "debug", false, "Use debug layers")
	flag.IntVar(&wwApp.pdIndex, "dev", 0, "Physical device index")
	ctx := initContext{}
	err := initApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize render application: ", err)
	}
	http.HandleFunc("/", doIndex)
	http.HandleFunc("/getImg", sendImage)
	fmt.Println("Running on localhost:3050")
	err = http.ListenAndServe(":3050", nil)
	if err != nil {
		log.Fatal("Start server failed ", err)
	}
}

func doIndex(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("content-type", "text/html")
	_, _ = writer.Write([]byte(indexPage))
}

type reqContext struct {
	writer http.ResponseWriter
}

func (r reqContext) SetError(err error) {
	r.writer.WriteHeader(500)
	r.writer.Write([]byte(err.Error()))
	panic("Failed")
}

func (r reqContext) IsValid() bool {
	return true
}

func (r reqContext) Begin(callName string) (atEnd func()) {
	return nil
}

func sendImage(writer http.ResponseWriter, request *http.Request) {
	uri, err := url.Parse(request.RequestURI)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	sAngle := uri.Query().Get("angle")
	angle, err := strconv.ParseFloat(sAngle, 64)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	ctx := reqContext{writer: writer}
	defer func() {
		_ = recover()
	}()

	s := image.Pt(1024, 768)
	// Render raw image to memory
	pngImage := renderImage(ctx, angle*math.Pi/180, s)

	// Write image out
	writer.Header().Add("Content-type", "image/png")
	writer.Write(pngImage)
}
