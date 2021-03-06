package flow_test

import (
	"strings"

	"github.com/laughlinez/flow"
	_ "github.com/laughlinez/flow/gadgets"
)

func ExampleNewCircuit() {
	g := flow.NewCircuit()
	g.Run()
}

func ExampleTransformer() {
	upper := flow.Transformer(func(m flow.Message) flow.Message {
		return strings.ToUpper(m.(string))
	})

	g := flow.NewCircuit()
	g.AddCircuitry("u", upper)
	g.Feed("u.In", "abc")
	g.Feed("u.In", "def")
	g.Run()
	// Output:
	// Lost string: ABC
	// Lost string: DEF
}

// func ExampleRunner() {
// 	desc := "foo: bar\nInput: In \nOutput:  Out\n\nhaha\nyes!"
// 	upper := flow.Runner(desc, func(in flow.Input, out flow.Output) {
// 		out.Send(strings.ToUpper((<-in).(string)))
// 	})
//
// 	g := flow.NewCircuit()
// 	g.AddCircuitry("u", upper)
// 	g.Feed("u.In", "abc")
// 	g.Feed("u.In", "def")
// 	g.Run()
// 	// Output:
// 	// Lost string: ABC
// 	// Lost string: DEF
// }

func ExampleCircuit_Label() {
	// new circuit to repeat each incoming message three times
	wg := flow.NewCircuit()
	wg.Add("r", "Repeater")
	wg.Feed("r.Num", 3)
	wg.Label("MyIn", "r.In")
	wg.Label("MyOut", "r.Out")

	g := flow.NewCircuit()
	g.AddCircuitry("wg", wg)
	g.Feed("wg.MyIn", "abc")
	g.Feed("wg.MyIn", "def")

	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: abc
	// Lost string: abc
	// Lost string: def
	// Lost string: def
	// Lost string: def
}

func ExampleNestedCircuit() {
	g1 := flow.NewCircuit()
	g1.Add("p", "Pipe")
	g1.Label("In", "p.In")
	g1.Label("Out", "p.Out")

	g2 := flow.NewCircuit()
	g2.Add("p", "Pipe")
	g2.Label("In", "p.In")
	g2.Label("Out", "p.Out")

	g3 := flow.NewCircuit()
	g3.AddCircuitry("g1", g1)
	g3.AddCircuitry("g2", g2)
	g3.Connect("g1.Out", "g2.In", 0)
	g3.Label("In", "g1.In")
	g3.Label("Out", "g2.Out")

	g := flow.NewCircuit()
	g.Add("p1", "Pipe")
	g.AddCircuitry("g", g3)
	g.Add("p2", "Pipe")
	g.Connect("p1.Out", "g.In", 0)
	g.Connect("g.Out", "p2.In", 0)
	g.Feed("p1.In", "abc")
	g.Feed("p1.In", "def")
	g.Feed("p1.In", "ghi")
	g.Run()
	// Output:
	// Lost string: abc
	// Lost string: def
	// Lost string: ghi
}
