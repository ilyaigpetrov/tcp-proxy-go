package main

import (
  "fmt"
  "syscall"
  "strings"

  "github.com/chifflier/nfqueue-go/nfqueue"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

func runInput(payload *nfqueue.Payload) int {

  fmt.Println("New pocket!")

  data := payload.Data

  var dst string
  var src string

  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

  var ip *layers.IPv4
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ = ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
      dst = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      src = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.SrcIP)), "."), "[]")
  }

  // var srcPort string
  // var dstPort string
  // Get the TCP layer from this packet
  var tcp *layers.TCP
  if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
      fmt.Println("This is a TCP packet!")
      // Get actual TCP data from this layer

      tcp, _ = tcpLayer.(*layers.TCP)

      tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())

      fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
      // dstPort = fmt.Sprintf("%d", tcp.DstPort)
      // srcPort = fmt.Sprintf("%d", tcp.SrcPort)

      dst = fmt.Sprintf("%s:%d", dst, tcp.DstPort)
      src = fmt.Sprintf("%s:%d", src, tcp.SrcPort)

  } else {
    fmt.Println("NOT TCP!")
    payload.SetVerdict(nfqueue.NF_ACCEPT)
    return 0
  }

  buf := gopacket.NewSerializeBuffer()
  opts := gopacket.SerializeOptions{
    ComputeChecksums: true,
  }
  gopacket.SerializeLayers(buf, opts,
    &layers.Ethernet{},
    ip,
    tcp,
    gopacket.Payload([]byte{1, 2, 3, 4}))
  packetData := buf.Bytes()




  err := gopacket.SerializeTo(pokBuf, opts)
  if err != nil {
    panic(err)
  }

  fmt.Printf("Proxying %s\n", dst)

  payload.SetVerdictModified(nfqueue.NF_ACCEPT, pokBuf)

  //dstAddr, err := net.ResolveTCPAddr("tcp", dst)
  //if err != nil {
  //  panic(err)
  //}
  //srcAddr, err := net.ResolveTCPAddr("tcp", src)
  //if err != nil {
  //  panic(err)
  //}
  //fmt.Printf("Connection to %s\n", dstAddr)
  //remote, err := net.DialTCP("tcp", nil, dstAddr)
  //if err != nil {
  //  panic(err)
  //}
  // defer remote.Close()
  //remote.Write(tcpBuf.Bytes())

  return 0

  //payload.SetVerdict(nfqueue.NF_STOP)
  //return 0

}

func createInputQueue() {

  q := new(nfqueue.Queue)
  defer q.Close()

  q.SetCallback(runInput)

  q.Init()

  q.Unbind(syscall.AF_INET)
  q.Bind(syscall.AF_INET)

  q.CreateQueue(13)

  q.Loop()
  q.DestroyQueue()

}

func main() {

  fmt.Println("Starting server...")
  go createInputQueue()
  fmt.Println("Waiting for queque 13.")
  select{}

}
