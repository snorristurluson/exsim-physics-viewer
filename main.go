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
	"time"
	"math"
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
	InRange [] int64
}

type State struct {
	Ships map[string] Ship
}

type SolarsystemViewer struct {
	state State
}

func NewSolarsystemViewer() (*SolarsystemViewer){
	return &SolarsystemViewer{}
}

func (viewer* SolarsystemViewer) render(imd* imdraw.IMDraw, atlas* text.Atlas, thickness float64) ([]*text.Text) {
	labels := []*text.Text{}
	shipsById := make(map[int64]Ship)
	for _, ship := range(viewer.state.Ships) {
		shipsById[ship.Owner] = ship
	}

	for _, ship := range(viewer.state.Ships) {
		x := ship.Position.X
		y := ship.Position.Y
		pos := pixel.V(x,y)

		imd.Color = colornames.Black
		imd.Push(pos)
		imd.Circle(10, thickness)

		if false {
			imd.Color = colornames.Gray
			imd.Push(pos)
			imd.Circle(100, thickness)

		}
		for _, shipInRange := range(ship.InRange) {
			other := shipsById[shipInRange]
			otherPos := pixel.V(other.Position.X, other.Position.Y)
			imd.Color = colornames.Gray
			imd.Push(pos, otherPos)
			imd.Line(thickness)
		}

		label := text.New(pos, atlas)
		label.Color = colornames.Black
		fmt.Fprintf( label,"%v", ship.Owner)
		labels = append(labels, label)
	}

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
	camPos := pixel.ZV
	camSpeed := 500.0
	camZoom := 1.0
	camZoomSpeed := 1.2

	imd := imdraw.New(nil)
	labels := []*text.Text{}

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII, text.RangeTable(unicode.Latin))
	numShips := text.New(pixel.V(10, 10), atlas)
	numShips.Color = colornames.Darkgreen
	fmt.Fprintf( numShips, "Ships: %v", len(viewer.state.Ships))

	recvChannel := make(chan State)
	go receive_loop(conn, recvChannel)
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		select {
		case stateReceived := <- recvChannel:
			viewer.state = stateReceived
			numShips.Clear()
			numShips.Dot = numShips.Orig
			fmt.Fprintf( numShips, "Ships: %v", len(viewer.state.Ships))
			imd.Clear()
			labels = viewer.render(imd, atlas, 1.0 / camZoom)
		default:
			// No data received
		}
		win.Clear(colornames.Skyblue)

		if win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		imd.Draw(win)
		for _, label := range(labels) {
			label.Draw(win, pixel.IM)
		}

		win.SetMatrix(pixel.IM)
		numShips.Draw(win, pixel.IM)

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