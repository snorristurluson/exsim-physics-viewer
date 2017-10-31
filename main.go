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
	"strings"
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
	return labels
}

func run() {
	//x := `{"command": "addship", "owner": 1, "type": 2, "position": {"x": 100.0, "y": 100.0, "z": 0.0}}"`
	//x := `{"command": "addship", "owner": 1, "type": 2, "position": {"x": 200.0, "y": 100.0, "z": 0.0}}"`
	//x := `{"command": "addship", "owner": 1, "type": 2, "position": {"x": 200.0, "y": 150.0, "z": 0.0}}"`
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
	basicTxt := text.New(pixel.V(10, 100), atlas)

	fmt.Fprintln(basicTxt, "Hello, text!")
	fmt.Fprintln(basicTxt, "I support multiple lines!")
	fmt.Fprintf(basicTxt, "And I'm an %s, yay!", "io.Writer")

	recvChannel := make(chan string)
	go receive_loop(conn, recvChannel)
	for !win.Closed() {
		select {
		case msgReceived := <- recvChannel:
			fmt.Printf("Message Received: %s\n", msgReceived)
			if !strings.HasPrefix(msgReceived, "error:") {
				json.Unmarshal([]byte(msgReceived), &viewer.state)
				imd.Clear()
				labels = viewer.render(imd, atlas)
			}
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

func receive_loop(conn net.Conn, c chan string) {
	recvBuf := make([]byte, 4096)
	for {
		n, err := conn.Read(recvBuf)
		if err != nil {
			fmt.Print(err)
		}
		msgReceived := string(recvBuf[:n])
		c <- msgReceived
	}
}

func main() {
	pixelgl.Run(run)
}