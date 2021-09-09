package main

import (
	"errors"
	"fmt"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/base"
	"io"
	"log"
	"net/http"
)

const pluginName = "debughttp"
const elementName = "debughttpsrc"

func httpFactoryFn(p *gst.Plugin) bool {
	return gst.RegisterElement(
		p,
		elementName,
		512,
		&httpSrc{},
		base.ExtendsBaseSrc,
		gst.InterfaceURIHandler,
	)
}

func RegisterHttpPlugin() (factoryName string, err error) {
	meta := &gst.PluginMetadata{
		MajorVersion: gst.VersionMajor,
		MinorVersion: gst.VersionMinor,
		Name:         pluginName,
		Description:  "HTTP Client with debug capabilities",
		Version:      "v0.0.1",
		License:      gst.LicenseLGPL,
		Source:       "debug-http",
		Package:      "Debug",
		Origin:       "https://github.com/cognitive-i/go-gst",
		ReleaseDate:  "2021-08-17",
	}

	if gst.RegisterPlugin(meta, httpFactoryFn) {
		return elementName, nil
	}
	return "", errors.New("unable to load")
}

type httpSrc struct {
	location string
	response *http.Response
}

func (g *httpSrc) New() glib.GoObjectSubclass {
	return &httpSrc{}
}

var properties = []*glib.ParamSpec{
	glib.NewStringParam(
		"location",                          // The name of the parameter
		"Location",                          // The long name for the parameter
		"Location of the file to read from", // A blurb about the parameter
		nil,                                 // A default value for the parameter
		glib.ParameterReadWrite,             // Flags for the parameter
	),
}

func (g *httpSrc) ClassInit(klass *glib.ObjectClass) {
	class := gst.ToElementClass(klass)
	class.SetMetadata(
		"Debug HTTP Source",
		"Source/Network",
		"Read stream from an HTTP server and snoop while the pipeline is doing its thing",
		"Cognitive-i Ltd",
	)
	class.AddPadTemplate(gst.NewPadTemplate(
		"src",
		gst.PadDirectionSource,
		gst.PadPresenceAlways,
		gst.NewAnyCaps(),
	))

	class.InstallProperties(properties)
}

func (g *httpSrc) setLocation(path string) error {
	if g.response != nil {
		return errors.New("cannot change location whilst running")
	}

	g.location = path
	return nil
}

func (g *httpSrc) SetProperty(self *glib.Object, id uint, value *glib.Value) {
	param := properties[id]
	switch param.Name() {
	case "location":
		var val string
		if value == nil {
			val = ""
		} else {
			val, _ = value.GetString()
		}
		if err := g.setLocation(val); err != nil {
			gst.ToElement(self).ErrorMessage(gst.DomainLibrary, gst.LibraryErrorSettings,
				fmt.Sprintf("Could not set location on object: %s", err.Error()),
				"",
			)
			return
		}
	}
}

func (g *httpSrc) GetProperty(self *glib.Object, id uint) *glib.Value {
	param := properties[id]
	switch param.Name() {
	case "location":
		if g.location == "" {
			return nil
		}
		val, err := glib.GValue(g.location)
		if err == nil {
			return val
		}
		gst.ToElement(self).ErrorMessage(gst.DomainLibrary, gst.LibraryErrorFailed,
			fmt.Sprintf("Could not convert %s to GValue", g.location),
			err.Error(),
		)
	}
	return nil
}

func (g *httpSrc) Start(*base.GstBaseSrc) bool {
	if g.response != nil {
		return false
	}

	if response, err := http.Get(g.location); err == nil {
		g.response = response
		return true
	} else {
		log.Print(err)
		return false
	}
}

func (g *httpSrc) IsSeekable(_ *base.GstBaseSrc) bool {
	return false
}

func (g *httpSrc) Fill(_ *base.GstBaseSrc, offset uint64, size uint, buffer *gst.Buffer) gst.FlowReturn {
	if g.response == nil {
		return gst.FlowError
	}

	if offset != 0 {
		return gst.FlowError
	}

	bufmap := buffer.Map(gst.MapWrite)
	if bufmap == nil {
		return gst.FlowError
	}
	defer buffer.Unmap()

	n, err := io.CopyN(bufmap.Writer(), g.response.Body, int64(size))
	buffer.SetSize(n)

	if err == io.EOF {
		if 0 == n {
			return gst.FlowEOS
		}
		return gst.FlowOK
	}

	if err != nil {
		return gst.FlowError
	} else {
		return gst.FlowOK
	}
}

func (g *httpSrc) GetURI() string {
	return g.location
}

func (g *httpSrc) GetURIType() gst.URIType {
	return gst.URISource
}

func (g *httpSrc) GetProtocols() []string {
	return []string{"http", "https"}
}

func (g *httpSrc) SetURI(s string) (bool, error) {
	err := g.setLocation(s)
	return err == nil, err
}
