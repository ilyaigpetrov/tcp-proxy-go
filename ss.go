// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package main

import (
  "io"
  "net"
  "sync"
  "fmt"
  "flag"
  "strings"

  log "github.com/Sirupsen/logrus"
  "github.com/google/gopacket"
  "github.com/google/gopacket/layers"
)

type Proxy struct {
  from string
  fromTCP *net.TCPAddr
  done chan struct{}
  log  *log.Entry
}

func NewProxy(from string) *Proxy {

  log.SetLevel(log.InfoLevel)
  return &Proxy{
    from: from,
    done: make(chan struct{}),
    log: log.WithFields(log.Fields{
      "from": from,
    }),
  }

}

func (p *Proxy) Start() error {
  p.log.Infoln("Starting proxy")
  var err error
  p.fromTCP, err = net.ResolveTCPAddr("tcp", p.from)
  if (err != nil) {
    panic(err)
  }
  listener, err := net.ListenTCP("tcp", p.fromTCP)
  if err != nil {
    return err
  }
  go p.run(*listener)
  return nil
}

func (p *Proxy) Stop() {
  p.log.Infoln("Stopping proxy")
  if p.done == nil {
    return
  }
  close(p.done)
  p.done = nil
}

func (p *Proxy) run(listener net.TCPListener) {
  for {
    select {
    case <-p.done:
      return
    default:
      connection, err := listener.AcceptTCP()
      if connection == nil {
        p.log.WithField("err", err).Errorln("Nil connection")
        panic(err)
      }
      la := connection.LocalAddr()
      if (la == nil) {
        panic("Connection lost!")
      }
      //fmt.Printf("Connection from %s\n", la.String())

      if err == nil {
        go p.handle(*connection)
      } else {
        p.log.WithField("err", err).Errorln("Error accepting conn")
      }
    }
  }
}

func (p *Proxy) handle(connection net.TCPConn) {

  defer connection.Close()
  p.log.Debugln("Handling", connection)
  defer p.log.Debugln("Done handling", connection)

  data := make([]byte, 0, 4096) // big buffer
  tmp := make([]byte, 256)     // using small tmo buffer for demonstrating
  for {
      n, err := connection.Read(tmp)
      if err != nil {
        if err != io.EOF {
              fmt.Println("read error:", err)
          }
          break
      }
      fmt.Println("got", n, "bytes.")
      data = append(data, tmp[:n]...)
  }


  var dest string
  // Decode a packet
  packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
  // Get the TCP layer from this packet
  ipLayer := packet.Layer(layers.LayerTypeIPv4)
  if ipLayer != nil {
      // Get actual TCP data from this layer
      ip, _ := ipLayer.(*layers.IPv4)
      fmt.Printf("From src %d to dst %d\n", ip.SrcIP, ip.DstIP)
      dest = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ip.DstIP)), "."), "[]")
  }
  fmt.Printf("Destination is: %s\n", dest)

  tcpBuf := gopacket.NewSerializeBuffer()
  opts := gopacket.SerializeOptions{
    ComputeChecksums: true,
  }  // See SerializeOptions for more details.

  // Get the TCP layer from this packet
  if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
      fmt.Println("This is a TCP packet!")
      // Get actual TCP data from this layer

      tcp, _ := tcpLayer.(*layers.TCP)

      tcp.SetNetworkLayerForChecksum(packet.NetworkLayer())

      fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
      dest = fmt.Sprintf("%s:%d", dest, tcp.DstPort)

      err := tcp.SerializeTo(tcpBuf, opts)
      if err != nil {
        panic(err)
      }
  } else {
    fmt.Println("NOT TCP!")
    return
  }
  fmt.Printf("Proxying %s\n", dest)


  /*
  var clientConn *net.TCPConn;
  ipv4, port, clientConn, err := getOriginalDst(&connection)
  if (err != nil) {
    panic(err)
  }
  connection = *clientConn;
  defer connection.Close()

  dest := ipv4 + ":" + fmt.Sprintf("%d", port)
  */

  if dest == *remoteAddr || dest == strings.Replace(*remoteAddr, "0.0.0.0", "127.0.0.1", -1) {
    fmt.Printf("DESTINATION IS SELF: %s", dest)
    return // NO SELF CONNECTIONS
  }

  addr, err := net.ResolveTCPAddr("tcp", dest)
  if err != nil {
    panic(err)
  }
  fmt.Printf("Connection to %s\n", dest)
  remote, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    p.log.WithField("err", err).Errorln("Error dialing remote host")
    return
  }
  defer remote.Close()
  remote.Write(tcpBuf.Bytes())

}

func (p *Proxy) copy(from, to net.TCPConn, wg *sync.WaitGroup) {
  defer wg.Done()
  select {
  case <-p.done:
    return
  default:
    if _, err := io.Copy(&to, &from); err != nil {
      p.log.WithField("err", err).Errorln("Error from copy")
      p.Stop()
      return
    }
  }
}

var remoteAddr *string = flag.String("r", "boom", "remote address")

func main() {

    flag.Parse();
    log.SetLevel(log.InfoLevel)

    NewProxy(*remoteAddr).Start()
    fmt.Println("Server started.")
    select{}
}
