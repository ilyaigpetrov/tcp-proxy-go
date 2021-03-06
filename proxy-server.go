package main
/*
#include <netinet/in.h>
#include <arpa/inet.h>
*/
import "C"

import (
  "io"
  "net"
  "sync"
  "fmt"
  "errors"
  "syscall"
  "flag"
  "golang.org/x/net/ipv4"
  "encoding/binary"

  log "github.com/Sirupsen/logrus"
)

const SO_ORIGINAL_DST = 80

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

func getOriginalDst(clientConn *net.TCPConn) (ipv4 string, port uint16, newTCPConn *net.TCPConn, err error) {
    if clientConn == nil {
        log.Debugf("copy(): oops, dst is nil!")
        err = errors.New("ERR: clientConn is nil")
        return
    }

    // test if the underlying fd is nil
    remoteAddr := clientConn.RemoteAddr()
    if remoteAddr == nil {
        log.Debugf("getOriginalDst(): oops, clientConn.fd is nil!")
        err = errors.New("ERR: clientConn.fd is nil")
        return
    }

    srcipport := fmt.Sprintf("%v", clientConn.RemoteAddr())

    newTCPConn = nil
    // net.TCPConn.File() will cause the receiver's (clientConn) socket to be placed in blocking mode.
    // The workaround is to take the File returned by .File(), do getsockopt() to get the original
    // destination, then create a new *net.TCPConn by calling net.TCPConn.FileConn().  The new TCPConn
    // will be in non-blocking mode.  What a pain.
    clientConnFile, err := clientConn.File()
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->FAILEDTOBEDETERMINED|ERR: could not get a copy of the client connection's file object", srcipport)
        return
    } else {
        clientConn.Close()
    }

    // Get original destination
    // this is the only syscall in the Golang libs that I can find that returns 16 bytes
    // Example result: &{Multiaddr:[2 0 31 144 206 190 36 45 0 0 0 0 0 0 0 0] Interface:0}
    // port starts at the 3rd byte and is 2 bytes long (31 144 = port 8080)
    // IPv4 address starts at the 5th byte, 4 bytes long (206 190 36 45)
    addr, err :=  syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    log.Debugf("getOriginalDst(): SO_ORIGINAL_DST=%+v\n", addr)
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->FAILEDTOBEDETERMINED|ERR: getsocketopt(SO_ORIGINAL_DST) failed: %v", srcipport, err)
        return
    }
    newConn, err := net.FileConn(clientConnFile)
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->%v|ERR: could not create a FileConn fron clientConnFile=%+v: %v", srcipport, addr, clientConnFile, err)
        return
    }
    if _, ok := newConn.(*net.TCPConn); ok {
        newTCPConn = newConn.(*net.TCPConn)
        clientConnFile.Close()
    } else {
        errmsg := fmt.Sprintf("ERR: newConn is not a *net.TCPConn, instead it is: %T (%v)", newConn, newConn)
        log.Infof("GETORIGINALDST|%v->?->%v|%s", srcipport, addr, errmsg)
        err = errors.New(errmsg)
        return
    }

    ipv4 = itod(uint(addr.Multiaddr[4])) + "." +
           itod(uint(addr.Multiaddr[5])) + "." +
           itod(uint(addr.Multiaddr[6])) + "." +
           itod(uint(addr.Multiaddr[7]))
    port = uint16(addr.Multiaddr[2]) << 8 + uint16(addr.Multiaddr[3])

    return
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
      fmt.Printf("Connection from %s\n", la.String())

      if err == nil {
        go p.handle(*connection)
      } else {
        p.log.WithField("err", err).Errorln("Error accepting conn")
      }
    }
  }
}



func ip2int(ip net.IP) uint32 {
  if len(ip) == 16 {
    return binary.BigEndian.Uint32(ip[12:16])
  }
  return binary.BigEndian.Uint32(ip)
}

func (p *Proxy) handle(connection net.TCPConn) {

  defer connection.Close()
  p.log.Debugln("Handling", connection)
  defer p.log.Debugln("Done handling", connection)

  buf := make([]byte, 0, 8186) // big buffer
  tmp := make([]byte, 4096)     // using small tmo buffer for demonstrating
  for {
      n, err := connection.Read(tmp)
      if err != nil {
        if err != io.EOF {
              fmt.Println("read error:", err)
          }
          break
      }
      fmt.Println("got", n, "bytes.")
      buf = append(buf, tmp[:n]...)
      header, err := ipv4.ParseHeader(buf)
      if err != nil {
        fmt.Println("Couldn't parse packet, dropping connnection.")
        return
      }
      if (header.TotalLen > len(buf)) {
        fmt.Println("Reading more up to %d\n", header.TotalLen)
        continue
      }
      packetData := buf[0:header.TotalLen]
      fmt.Printf("PACKET LEN:%d, bufLen:%d\n", header.TotalLen, len(buf))
      buf = buf[header.TotalLen:]
      fmt.Printf("Packet to %s\n", header.Dst)


      s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
      if err != nil {
        fmt.Printf("Error", err)
        return
      }


      var arr [4]byte
      copy(arr[:], header.Dst.To4()[:4])
      addr := syscall.SockaddrInet4{
        Addr: arr,
      }

      //addr := ip2int(header.Dst)

      //sin.sin_addr = C.struct_in_addr{ s_addr: addr }
      syscall.Sendto (s, packetData, 0, &addr)

  }


  /*

  dest := ipv4 + ":" + fmt.Sprintf("%d", port)
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
  wg := &sync.WaitGroup{}
  wg.Add(2)
  go p.copy(*remote, connection, wg)
  go p.copy(connection, *remote, wg)
  wg.Wait()
  */

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

func itod(i uint) string {
        if i == 0 {
                return "0"
        }

        // Assemble decimal in reverse order.
        var b [32]byte
        bp := len(b)
        for ; i > 0; i /= 10 {
                bp--
                b[bp] = byte(i%10) + '0'
        }

        return string(b[bp:])
}

var remoteAddr *string = flag.String("r", "boom", "remote address")

func main() {

    flag.Parse();
    log.SetLevel(log.InfoLevel)

    NewProxy(*remoteAddr).Start()
    fmt.Println("Server started.")
    select{}
}
