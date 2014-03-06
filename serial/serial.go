// Interface to serial port devices.
package serial

import (
	"bufio"
	"strings"
	"time"

	"github.com/chimera/rs232"
	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["TimeStamp"] = func() flow.Worker { return &TimeStamp{} }
	flow.Registry["SerialIn"] = func() flow.Worker { return &SerialIn{} }
	flow.Registry["SketchType"] = func() flow.Worker { return &SketchType{} }
}

// Insert a timestamp before each message. Registers as "TimeStamp".
type TimeStamp struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start inserting timestamps.
func (w *TimeStamp) Run() {
	for m := range w.In {
		w.Out.Send(time.Now())
		w.Out.Send(m)
	}
}

// Line-oriented serial input port, opened once the Port input is set.
type SerialIn struct {
	flow.Work
	Port flow.Input
	Out  flow.Output
}

// Start processing incoming text lines from the serial interface.
// Registers as "SerialIn".
func (w *SerialIn) Run() {
	if port, ok := <-w.Port; ok {
		opt := rs232.Options{BitRate: 57600, DataBits: 8, StopBits: 1}
		dev, err := rs232.Open(port.(string), opt)
		check(err)

		scanner := bufio.NewScanner(dev)
		for scanner.Scan() {
			w.Out.Send(scanner.Text())
		}
	}
}

// SketchType looks for lines of the form "[name...]" in the input stream.
// These then cause a corresponding worker to be loaded dynamically.
// Registers as "SketchType".
type SketchType struct {
	flow.Work
	In     flow.Input
	ViaOut flow.Output // send to dynamically added worker
	ViaIn  flow.Input  // receive from dynamically added worker
	Out    flow.Output
}

// Start transforming the "[name...]" markers in the input stream.
func (w *SketchType) Run() {
	for m := range w.In {
		if s, ok := m.(string); ok {
			if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
				tag := "Sketch-" + s[1:strings.IndexAny(s, ".]")]
				// FIXME: this code is a horrible hack
				if _, ok := flow.Registry[tag]; ok {

					g := w.MyGroup()
					g.Add("(sketch)", tag)
					g.Connect(w.MyName()+".ViaOut", "(sketch).In", 0)
					g.Connect("(sketch).Out", w.MyName()+".ViaIn", 0)
					g.Launch("(sketch)")

					// start extra goroutine to copy ViaIn to Out
					go func() {
						for m := range w.ViaIn {
							w.Out.Send(m)
						}
					}()
				}
			}
		}
		w.ViaOut.Send(m)
	}
}

// Generic error checking, panics if e is not nil.
func check(e error) {
	if e != nil {
		panic(e)
	}
}
