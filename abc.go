// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package main

import (
  "net"
  "fmt"
  "syscall"
  "flag"
  "sync"

  log "github.com/Sirupsen/logrus"
  "github.com/chifflier/nfqueue-go/nfqueue"

  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

type Proxy struct {
  to string
  log  *log.Entry
}

func NewProxy(to string) *Proxy {

  log.SetLevel(log.InfoLevel)
  return &Proxy{
    to: to,
    log: log.WithFields(log.Fields{
      "to": to,
    }),
  }

}

func (p *Proxy) Start() error {
  p.log.Infoln("Starting proxy")


  q := new(nfqueue.Queue)

  q.SetCallback(run)

  q.Init()

  q.Unbind(syscall.AF_INET)
  q.Bind(syscall.AF_INET)

  q.CreateQueue(13)

  q.Loop()
  q.DestroyQueue()
  q.Close()

  return nil
}

func run(payload *nfqueue.Payload) int {

  fmt.Println("run")

  // Decode a packet
  packet := gopacket.NewPacket(payload.Data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src port %d to dst port %d\n", ip.SrcIP, ip.DstIP)
  }

  handle(payload.Data)
  return nfqueue.NF_DROP

}

var mutex = &sync.Mutex{}

func handle(data []byte) {

  fmt.Println("handle")
  toTCP, err := net.ResolveTCPAddr("tcp", p.to)
  if err != nil {
    panic(nil)
  }
  fmt.Println("dial")
  remote, err := net.DialTCP("tcp", nil, toTCP)
  if err != nil {
    p.log.WithField("err", err).Errorln("Error dialing remote host")
    return
  }
  defer remote.Close()

  fmt.Println("lock")
  mutex.Lock()
  fmt.Println("write data...")
  bw, err := remote.Write(data)
  mutex.Unlock()
  if err != nil {
    panic(err)
  }
  if bw != len(data) {
    panic(fmt.Sprintf("Not all data written: %s/%s", bw, len(data)))
  }

}

var remoteAddr *string = flag.String("r", "boom", "remote address")

var p *Proxy

func main() {

    flag.Parse();
    log.SetLevel(log.InfoLevel)
    if *remoteAddr == "boom" {
      panic("Specify proxy server address!")
    }

    p = NewProxy(*remoteAddr)
    p.Start()
    fmt.Println("Server started.")
    select{}
}
