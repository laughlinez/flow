package flow

import (
	"reflect"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// Initialise a new circuit.
func NewCircuit() *Circuit {
	c := Circuit{
		gadgets: map[string]*Gadget{},
		wires:   map[string]int{},
		feeds:   map[string][]Message{},
		labels:  map[string]string{},
		null:    make(chan Message),
	}
	close(c.null)
	return &c
}

// A circuit is a collection of inter-connected gadgets.
type Circuit struct {
	Gadget

	gadgets map[string]*Gadget   // gadgets added to this circuit
	wires   map[string]int       // all wire definitions
	feeds   map[string][]Message // all message feeds
	labels  map[string]string    // pin label lookup map

	null chan Message // used for dangling inputs
	sink chan Message // used for dangling outputs

	wait sync.WaitGroup // tracks number of running gadgets
}

func (c *Circuit) initPins() {
	// fill c.inputs[]
	// fill c.outputs[]
	glog.Errorln("c-initpins", c.name)
}

// Add a named gadget to the circuit with a unique name.
func (c *Circuit) Add(name, gadget string) {
	constructor := Registry[gadget]
	if constructor == nil {
		glog.Errorln("not found:", gadget)
		return
	}
	g := c.AddCircuitry(name, constructor())
	g.regType = gadget
}

// Add a gadget or circuit to the circuit with a unique name.
func (c *Circuit) AddCircuitry(name string, circ Circuitry) *Gadget {
	g := circ.initGadget(circ, name, c)
	c.gadgets[name] = g
	circ.initPins()
	return g
}

// Connect an output pin with an input pin.
func (c *Circuit) Connect(from, to string, capacity int) {
	// c.wires = append(c.wires, wireDef{from, to, capacity})
	// w := c.gadgetOf(to).getInput(pinPart(to), capacity)
	// c.gadgetOf(from).setOutput(pinPart(from), w)
	c.wires[from+"/"+to] = capacity
}

// Set up a message to feed to a gadget on startup.
func (c *Circuit) Feed(pin string, m Message) {
	c.feeds[pin] = append(c.feeds[pin], m)
}

// Label an external pin to map it to an internal one.
// TODO: get rid of this, use wires with undotted in or out pins instead
func (c *Circuit) Label(external, internal string) {
	if strings.Contains(external, ".") {
		glog.Fatalln("external pin should not include a dot:", external)
	}
	c.labels[external] = internal
}

type wire struct {
	fanIn   int
	channel chan Message
}

// Start up the circuit, and return when it is finished.
func (c *Circuit) Run() {
	inbound := map[string]*wire{}
	outbound := map[string]*wire{}

	glog.Errorln("feeds", c.name, len(c.feeds))
	for k, v := range c.feeds {
		inbound[k] = &wire{channel: make(chan Message, len(v))}
	}

	glog.Errorln("wires", c.name, len(c.wires))
	for wpair, wcap := range c.wires {
		v := strings.Split(wpair, "/")
		from := v[0]
		to := v[1]

		if _, ok := inbound[to]; !ok {
			inbound[to] = &wire{}
		}
		in := inbound[to]
		if cap(in.channel) < wcap {
			in.channel = make(chan Message, wcap) // replace with larger one
		}
		outbound[from] = in
	}

	glog.Errorln("fill", c.name)
	for k, v := range c.feeds {
		for _, f := range v {
			inbound[k].channel <- f
		}
	}

	glog.Errorln("sink", c.name)
	c.sink = make(chan Message)
	go func() {
		for m := range c.sink {
			glog.Errorln("lost:", c.name, m)
		}
	}()

	glog.Errorln("gadgets", c.name, len(c.wires))
	for _, g := range c.gadgets {
		c.wait.Add(1)

		glog.Errorln("g-in", g.name, len(g.inputs))
		for k, v := range g.inputs {
			if in, ok := inbound[k]; ok {
				setPin(v, in.channel)
			} else {
				setPin(v, c.null) // feed eof to unconnected inputs
			}
		}

		glog.Errorln("g-out", g.name, len(g.outputs))
		for k, v := range g.outputs {
			if out, ok := outbound[k]; ok {
				out.fanIn++
				setPin(v, out.channel)
			} else {
				setPin(v, c.sink) // ignore data from unconnected outputs
			}
		}

		glog.Errorln("g-close", g.name)
		for _, in := range inbound {
			if in.fanIn == 0 {
				close(in.channel)
			}
		}

		glog.Errorln("g-go", g.name)
		go func() {
			defer DontPanic()
			defer c.wait.Done()
			defer c.teardownPins(g)

			glog.Errorln("g-run", g.name)
			g.circuitry.Run()
			glog.Errorln("g-end", g.name)
		}()
	}
	c.wait.Wait()
	glog.Errorln("g-done", c.name)
	close(c.sink)
	c.sink = nil // this also marks the circuit as not running
	// clean up all channels
}

func (c *Circuit) teardownPins(g *Gadget) {
}

func setPin(v reflect.Value, c chan Message) {
	v.Set(reflect.ValueOf(c))
}
