package main

import (
	"context"
	"errors"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"log"
	"time"
)

var NilValueError = errors.New("unexpected nil value")

func throw(err error) {
	log.Panic(err)
}

func throwIf(errMaybe interface{}, parameters ...interface{}) {
	if l := len(parameters); l >= 1 {
		if errMaybe == nil {
			throw(NilValueError)
		}
		errMaybe = parameters[l-1]
	}

	if err, ok := errMaybe.(error); ok && (err != nil) {
		throw(err)
	}
}

func main() {
	gst.Init(nil)
	defer gst.Deinit()

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
	defer mainLoop.Quit()
	go mainLoop.Run()

	factoryName, err := RegisterHttpPlugin()
	throwIf(factoryName, err)

	explicitInvocation(factoryName)
	byUriInvocation(factoryName)
}

func explicitInvocation(factoryName string) {
	pipeline, err := gst.NewPipeline("static plugin example")
	throwIf(pipeline, err)

	elements, err := gst.NewElementMany(factoryName, "fakesink")
	throwIf(elements, err)
	throwIf(pipeline.AddMany(elements...))

	httpSource := elements[0]
	fakeSink := elements[1]

	throwIf(httpSource.SetProperty("location", "https://www.example.com"))
	throwIf(fakeSink.SetProperty("dump", true))
	throwIf(httpSource.Link(fakeSink))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		if msg.Type() == gst.MessageEOS {
			cancel()
			return false
		}
		return true
	})

	throwIf(pipeline.SetState(gst.StatePlaying))
	defer func() {
		throwIf(pipeline.SetState(gst.StateNull))
	}()

	<-ctx.Done()
	if err := ctx.Err(); err == context.DeadlineExceeded {
		throw(err)
	}
}

func byUriInvocation(factoryName string) {
	// This is roughly the technique fancy bins work with typefind
	// For example if you wanted to inject a custom HTTP client into a dashdemux to peep into
	// HTTPS traffic one way you could do it is with a static plugin and a high rank
	element, err := gst.NewElementFromUri(gst.URISource, "https://another.example.com", "httpSrc")
	throwIf(element, err)
	if actualFactory := element.GetFactory().GetName(); actualFactory != factoryName {
		throw(errors.New("unexpected element " + actualFactory))
	}
}
