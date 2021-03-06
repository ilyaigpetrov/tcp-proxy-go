package main

import (
  "net"
  "fmt"
  "syscall"
  "flag"
  "strings"

  "github.com/chifflier/nfqueue-go/nfqueue"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

func run(payload *nfqueue.Payload) int {

  fmt.Println("run")
  handle(payload.Data)
  fmt.Println("STOLLEN")
  payload.SetVerdict(2)
  return 0

}

func handle(data []byte) {

  fmt.Println("handle")
  toTCP, err := net.ResolveTCPAddr("tcp", *remoteAddr)
  if err != nil {
    panic(nil)
  }

  fmt.Println("dial")
  remote, err := net.DialTCP("tcp", nil, toTCP)
  if err != nil {
    panic(err)
  }
  defer remote.Close()

  //buf := gopacket.NewSerializeBuffer()
  //opts := gopacket.SerializeOptions{}  // See SerializeOptions for more details.

  var dest string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  ipLayer := packet.Layer(layers.LayerTypeIPv4)
  if ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
      dest = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      //err := tcp.SerializeTo(buf, opts)
      //if err != nil {
      //  panic(err)
      //}
  }

  // Decode a packet
  // Get the TCP layer from this packet
  if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
      fmt.Println("This is a TCP packet!")
      // Get actual TCP data from this layer
      tcp, _ := tcpLayer.(*layers.TCP)
      fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
      dest = fmt.Sprintf("%s:%d", dest, tcp.DstPort)
  } else {
    fmt.Println("NOT TCP")
    return
  }
  fmt.Printf("Proxying %s\n", dest)

  fmt.Println("write data...")

  b := data

  //fmt.Println("HEX", hex.Dump(b))
  wcount := 0
  for {
    wc, err := remote.Write(b)
    if err != nil {
      panic(err)
    }
    wcount += wc
    if wcount == len(b) {
      break
    }
  }


}

var remoteAddr *string = flag.String("r", "boom", "remote address")

var remote net.TCPConn

func main() {

    flag.Parse();
    if *remoteAddr == "boom" {
      panic("Specify proxy server address!")
    }
    fmt.Println("Starting server...")

    q := new(nfqueue.Queue)

    q.SetCallback(run)

    q.Init()

    q.Unbind(syscall.AF_INET)
    q.Bind(syscall.AF_INET)

    q.CreateQueue(14)

    q.Loop()
    q.DestroyQueue()
    q.Close()

}
