package main

import (
  "net"
  "fmt"
  "syscall"
  "strings"

  "github.com/chifflier/nfqueue-go/nfqueue"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

func run(payload *nfqueue.Payload) int {

  fmt.Println("run")
  handle(payload.Data)
  fmt.Println("DROP")
  return nfqueue.NF_ACCEPT

}

func handle(data []byte) {

  fmt.Println("handle")

  var dest string
  var src string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
      dest = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      src = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.SrcIP)), "."), "[]")
  }

  tcpBuf := gopacket.NewSerializeBuffer()
  opts := gopacket.SerializeOptions{}  // See SerializeOptions for more details.
  // var srcPort string
  // var dstPort string
  // Get the TCP layer from this packet
  if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
      fmt.Println("This is a TCP packet!")
      // Get actual TCP data from this layer

      tcp, _ := tcpLayer.(*layers.TCP)

      tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())

      fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
      // dstPort = fmt.Sprintf("%d", tcp.DstPort)
      // srcPort = fmt.Sprintf("%d", tcp.SrcPort)

      dest = fmt.Sprintf("%s:%d", dest, tcp.DstPort)
      src = fmt.Sprintf("%s:%d", src, tcp.SrcPort)
      err := tcp.SerializeTo(tcpBuf, opts)
      if err != nil {
        panic(err)
      }
  } else {
    fmt.Println("NOT TCP!")
    return
  }

  fmt.Printf("Proxying %s\n", dest)

  dstAddr, err := net.ResolveTCPAddr("tcp", dest)
  if err != nil {
    panic(err)
  }
  //srcAddr, err := net.ResolveTCPAddr("tcp", src)
  //if err != nil {
  //  panic(err)
  //}
  fmt.Printf("Connection to %s\n", dstAddr)
  remote, err := net.DialTCP("tcp", nil, dstAddr)
  if err != nil {
    panic(err)
  }
  // defer remote.Close()
  remote.Write(tcpBuf.Bytes())

}

func main() {

    fmt.Println("Starting server...")

    q := new(nfqueue.Queue)

    q.SetCallback(run)

    q.Init()

    q.Unbind(syscall.AF_INET)
    q.Bind(syscall.AF_INET)

    q.CreateQueue(13)

    q.Loop()
    q.DestroyQueue()
    q.Close()

}
