group = 
  workers: [
    { type: "SerialIn", name: "s" }
    { type: "TimeStamp", name: "t" }
    { type: "SketchType", name: "u" }
    { type: "Printer", name: "p" }
  ]
  connections: [
    { from: "s.Out", to: "t.In" }
    { from: "t.Out", to: "u.In" }
    { from: "u.Out", to: "p.In" }
  ]
  requests: [
    data: "/dev/tty.usbserial-A901ROSM", to: "s.Port"
  ]

console.log JSON.stringify group, null, 4
