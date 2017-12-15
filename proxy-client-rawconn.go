package main

import (
  "net"
  "fmt"
  "syscall"
  "flag"
  "sync"
  "strings"
  "os"

  "github.com/chifflier/nfqueue-go/nfqueue"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

func run(payload *nfqueue.Payload) int {

  fmt.Println("run")
  handle(payload.Data)
  return nfqueue.NF_DROP

}

var mutex = &sync.Mutex{}

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

  buf := gopacket.NewSerializeBuffer()
  opts := gopacket.SerializeOptions{}  // See SerializeOptions for more details.

  var dest string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
      dest = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
      err := ip.SerializeTo(buf, opts)
      if err != nil {
        panic(err)
      }
  }
  fmt.Printf("To: %s", dest)


  fmt.Println("lock")
  mutex.Lock()

  fmt.Println("write data...")
  b := buf.Bytes()
  syscon, err := remote.SyscallConn()
  if err != nil {
    panic(err)
  }
  err = syscon.Write(
    func(fd uintptr) (done bool) {

      file := os.NewFile(fd, "pipe")
      _, err := file.Write(b)
      if err != nil {
        panic(err)
      }
      return true

    },
  )
  if err != nil {
    panic(err)
  }

  mutex.Unlock()

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

    q.CreateQueue(13)

    q.Loop()
    q.DestroyQueue()
    q.Close()

}
