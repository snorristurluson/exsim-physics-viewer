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
	"github.com/faiface/pixel/imdraw"
)

type CommandResult struct {
	Result string
	State json.RawMessage
}

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

type SolarsystemViewer struct {
	state State
}

func NewSolarsystemViewer() (*SolarsystemViewer){
	return &SolarsystemViewer{}
}

func (viewer* SolarsystemViewer) render(imd* imdraw.IMDraw, atlas* text.Atlas) ([]*text.Text) {
	labels := []*text.Text{}
	for i, ship := range(viewer.state.Ships) {
		x := ship.Position.X
		y := ship.Position.Y
		imd.Color = colornames.Black
		imd.Push(pixel.V(x,y))
		imd.Circle(10, 1)
		label := text.New(pixel.V(x, y), atlas)
		label.Color = colornames.Black
		fmt.Fprintf( label,"%v", i)
		labels = append(labels, label)
	}

	numShips := text.New(pixel.V(10, 10), atlas)
	numShips.Color = colornames.Darkgreen
	fmt.Fprintf( numShips, "Ships: %v", len(viewer.state.Ships))
	labels = append(labels, numShips)

	return labels
}

func run() {
	full_address := "localhost:4041"
	conn, err := net.Dial("tcp", full_address)
	if err != nil {
		fmt.Errorf("Couldn't connect")
		return
	}
	fmt.Printf("Connection made\n")

	cfg := pixelgl.WindowConfig{
		Title:  "exsim-physics-viewer",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	viewer := NewSolarsystemViewer()
	imd := imdraw.New(nil)
	labels := []*text.Text{}

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII, text.RangeTable(unicode.Latin))

	recvChannel := make(chan State)
	go receive_loop(conn, recvChannel)
	for !win.Closed() {
		select {
		case stateReceived := <- recvChannel:
			viewer.state = stateReceived
			imd.Clear()
			labels = viewer.render(imd, atlas)
		default:
			// No data received
		}
		win.Clear(colornames.Skyblue)
		imd.Draw(win)
		for _, label := range(labels) {
			label.Draw(win, pixel.IM)
		}
		win.Update()
	}
}

func receive_loop(conn net.Conn, c chan State) {
	decoder := json.NewDecoder(conn)
	for {
		var cmd CommandResult
		err := decoder.Decode(&cmd)
		if err != nil {
			fmt.Printf("Error in Decode: %v\n", err)
			break
		}
		if cmd.Result == "state" {
			var state State
			err := json.Unmarshal(cmd.State, &state)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			c <- state
		}
	}
}

func main() {
	pixelgl.Run(run)
}