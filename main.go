package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"unicode"
	"golang.org/x/image/font/basicfont"
	"fmt"
	"net"
	"encoding/json"
)

type Vector3 struct {
	X float64
	Y float64
	Z float64
}
type Ship struct {
	Owner int64
	Type int64
	Position Vector3
}
type State struct {
	Ships [] Ship
}

func run() {
	recvBuf := make([]byte, 4096)
	full_address := "localhost:4041"
	conn, err := net.Dial("tcp", full_address)
	if err != nil {
		fmt.Errorf("Couldn't connect")
		return
	}
	fmt.Printf("Connection made\n")
	conn.Write([]byte("{\"command\": \"addship\", \"owner\": 1, \"type\": 2, \"position\": {\"x\": -10.0, \"y\": 10.0, \"z\": 0.0}}"))
	n, err := conn.Read(recvBuf)
	if err != nil {
		fmt.Print(err)
	}
	msgReceived := string(recvBuf[:n])
	fmt.Printf("Message Received: %s\n", msgReceived)

	conn.Write([]byte("{\"command\": \"getstate\"}"))
	n, err = conn.Read(recvBuf)
	if err != nil {
		fmt.Print(err)
	}
	msgReceived = string(recvBuf[:n])
	fmt.Printf("Message Received: %v\n", msgReceived)

	var state State
	err = json.Unmarshal(recvBuf[:n], &state)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%v", state)

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}


	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII, text.RangeTable(unicode.Latin))
	basicTxt := text.New(pixel.V(10, 100), atlas)

	fmt.Fprintln(basicTxt, "Hello, text!")
	fmt.Fprintln(basicTxt, "I support multiple lines!")
	fmt.Fprintf(basicTxt, "And I'm an %s, yay!", "io.Writer")

	for !win.Closed() {
		win.Clear(colornames.Skyblue)
		basicTxt.Draw(win, pixel.IM)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}