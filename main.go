package main

import (
	"fmt"
	"log"
	"syscall/js"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/htmlcanvas"
)

func renderWrapper() js.Func {
	renderFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		go func() {
			if len(args) != 1 {
				fmt.Println("Invalid no of arguments passed")
			}
			doc := js.Global().Get("document")
			if !doc.Truthy() {
				fmt.Println("Unable to get document object")
			}

			// Create a new logger that writes to the textarea
			logTextArea := doc.Call("getElementById", "log")
			if !logTextArea.Truthy() {
				fmt.Println("Unable to get log text area")
			}
			writer := jsWriter{logTextArea: logTextArea}
			logger := log.New(writer, "", 0)
			cvs := doc.Call("getElementById", "canvas")
			height := cvs.Get("height").Float()
			c := htmlcanvas.New(cvs, height, height, 1.0)
			ctx := canvas.NewContext(c)

			options := args[0].String()
			Render(ctx, options, logger)
		}()
		return nil
	})
	return renderFunc
}

// Define a custom writer
type jsWriter struct {
	logTextArea js.Value
}

// Implement the Write method for the io.Writer interface
func (jw jsWriter) Write(p []byte) (n int, err error) {
	// Append the text to the textarea, converting bytes to a JavaScript string
	currentText := jw.logTextArea.Get("value").String()
	newText := currentText + string(p)
	jw.logTextArea.Set("value", newText)

	// Return the number of bytes written and no error
	return len(p), nil
}

func main() {

	fmt.Println("Go WASM app loaded.")
	js.Global().Set("Render", renderWrapper())
	<-make(chan struct{})
}
