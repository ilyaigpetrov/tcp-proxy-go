package main

import (
  "net"
  "fmt"
  "syscall"
  "strings"
  "sort"
  "io"

  "github.com/chifflier/nfqueue-go/nfqueue"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

var ADDR_TO_CONN = make(map[[2]string]*net.TCPConn)

func toPair(slice []string) [2]string {

  var res [2]string
  sort.Strings(slice)
  copy(res[:], slice[:2])
  return res

}

func runInput(payload *nfqueue.Payload) int {

  fmt.Println("run inp")
  if handleInput(payload) != true {
    fmt.Println("ACCEPT")
    payload.SetVerdict(nfqueue.NF_ACCEPT)
  }
  return 0

}

func runOutput(payload *nfqueue.Payload) int {

  fmt.Println("run out")
  if handleOutput(payload) != true {
    fmt.Println("DROP")
    payload.SetVerdict(nfqueue.NF_DROP)
  }
  return 0

}

func handleInput(payload *nfqueue.Payload) bool {

  data := payload.Data
  var dst string
  var src string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      ip, _ := ipLayer.(*layers.IPv4)
      dst = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      src = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.SrcIP)), "."), "[]")
  }

  if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
      tcp, _ := tcpLayer.(*layers.TCP)
      dst = fmt.Sprintf("%s:%d", dst, tcp.DstPort)
      src = fmt.Sprintf("%s:%d", src, tcp.SrcPort)
  } else {
    fmt.Println("NOT TCP!")
    return false
  }
  arr := []string{src,dst}
  conn := ADDR_TO_CONN[toPair(arr)]
  if conn == nil {
    return false
  }
  // pack response to packet!
  payload.SetVerdictModified(nfqueue.NF_ACCEPT, payload.Data)
  return true

}

func handleOutput(payload *nfqueue.Payload) bool {

  fmt.Println("handle")
  data := payload.Data

  var dst string
  var src string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
      dst = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      src = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.SrcIP)), "."), "[]")
  }

  // var srcPort string
  // var dstPort string
  // Get the TCP layer from this packet
  tcpLayer := packet.Layer(layers.LayerTypeTCP)
  if tcpLayer != nil {
      fmt.Println("This is a TCP packet!")
      // Get actual TCP data from this layer

      tcp, _ := tcpLayer.(*layers.TCP)

      fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
      // dstPort = fmt.Sprintf("%d", tcp.DstPort)
      // srcPort = fmt.Sprintf("%d", tcp.SrcPort)

      dst = fmt.Sprintf("%s:%d", dst, tcp.DstPort)
      src = fmt.Sprintf("%s:%d", src, tcp.SrcPort)

  } else {
    fmt.Println("NOT TCP!")
    return false
  }

  fmt.Printf("Proxying %s\n", dst)

  dstAddr, err := net.ResolveTCPAddr("tcp", dst)
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
  remote.Write(tcpLayer.LayerPayload())
  // ADDR_TO_CONN[toPair([]string{src,dst})] = remote

  ans := make([]byte, 0, 4096) // big buffer
  tmp := make([]byte, 256)     // using small tmo buffer for demonstrating
  for {
      n, err := remote.Read(tmp)
      if err != nil {
        if err != io.EOF {
              fmt.Println("read error:", err)
          }
          break
      }
      fmt.Println("got", n, "bytes.")
      ans = append(ans, tmp[:n]...)
  }

  payload.SetVerdictModified(nfqueue.NF_ACCEPT, ans)
  return true

}

func createInputQueue() {

  q := new(nfqueue.Queue)

  q.SetCallback(runInput)

  q.Init()

  q.Unbind(syscall.AF_INET)
  q.Bind(syscall.AF_INET)

  q.CreateQueue(13)

  q.Loop()
  q.DestroyQueue()
  q.Close()

}

func createOutputQueue() {

  q := new(nfqueue.Queue)

  q.SetCallback(runOutput)

  q.Init()

  q.Unbind(syscall.AF_INET)
  q.Bind(syscall.AF_INET)

  q.CreateQueue(14)

  q.Loop()
  q.DestroyQueue()
  q.Close()

}

func main() {

  fmt.Println("Starting server...")
  // go createInputQueue()
  go createOutputQueue()
  select{}

}
