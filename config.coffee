group = 
  workers: [
    { name: "si", type: "SerialIn" }
    { name: "ts", type: "TimeStamp" }
    { name: "st", type: "SketchType" }
    { name: "nm", type: "Decoder-NodeMap" }
    # { name: "nm", type: "Pipe" }
    { name: "p", type: "Printer" }
  ]
  connections: [
    { from: "si.Out", to: "ts.In" }
    { from: "ts.Out", to: "st.In" }
    { from: "st.Out", to: "nm.In" }
    { from: "nm.Out", to: "p.In" }
  ]
  requests: [
    { data: "RFg5i2 roomNode", to: "nm.Info" }
    { data: "RFg5i3 radioBlip", to: "nm.Info" }
    { data: "RFg5i4 roomNode", to: "nm.Info" }
    { data: "RFg5i5 roomNode", to: "nm.Info" }
    { data: "RFg5i6 roomNode", to: "nm.Info" }
    { data: "RFg5i9 homePower", to: "nm.Info" }
    { data: "RFg5i10 roomNode", to: "nm.Info" }
    { data: "RFg5i11 roomNode", to: "nm.Info" }
    { data: "RFg5i12 roomNode", to: "nm.Info" }
    { data: "RFg5i13 roomNode", to: "nm.Info" }
    { data: "RFg5i15 smaRelay", to: "nm.Info" }
    { data: "RFg5i18 p1scanner", to: "nm.Info" }
    { data: "RFg5i19 ookRelay", to: "nm.Info" }
    { data: "RFg5i23 roomNode", to: "nm.Info" }
    { data: "RFg5i24 roomNode", to: "nm.Info" }
    
    { data: "/dev/tty.usbserial-A901ROSM", to: "si.Port" }
  ]

console.log JSON.stringify group, null, 4
