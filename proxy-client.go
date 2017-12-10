// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package main

import (
  "io"
  "net"
  "sync"
  "flag"
  "fmt"

  log "github.com/Sirupsen/logrus"
)

type Proxy struct {
  from string
  to   string
  done chan struct{}
  log  *log.Entry
}

func NewProxy(from, to string) *Proxy {
  return &Proxy{
    from: from,
    to:   to,
    done: make(chan struct{}),
    log: log.WithFields(log.Fields{
      "from": from,
      "to":   to,
    }),
  }
}

func (p *Proxy) Start() error {
  p.log.Infoln("Starting proxy")
  listener, err := net.Listen("tcp", p.from)
  if err != nil {
    return err
  }
  go p.run(listener)
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

func (p *Proxy) run(listener net.Listener) {
  for {
    select {
    case <-p.done:
      return
    default:
      connection, err := listener.Accept()
      if err == nil {
        go p.handle(connection)
      } else {
        p.log.WithField("err", err).Errorln("Error accepting conn")
      }
    }
  }
}

func (p *Proxy) handle(connection net.Conn) {
  p.log.Debugln("Handling", connection)

  fmt.Printf("Connection to %s\n", connection.RemoteAddr().String())

  defer p.log.Debugln("Done handling", connection)
  defer connection.Close()
  remote, err := net.Dial("tcp", p.to)
  if err != nil {
    p.log.WithField("err", err).Errorln("Error dialing remote host")
    return
  }
  defer remote.Close()
  wg := &sync.WaitGroup{}
  wg.Add(2)
  go p.copy(remote, connection, wg)
  go p.copy(connection, remote, wg)
  wg.Wait()
}

func (p *Proxy) copy(from, to net.Conn, wg *sync.WaitGroup) {
  defer wg.Done()
  select {
  case <-p.done:
    return
  default:
    if _, err := io.Copy(to, from); err != nil {
      p.log.WithField("err", err).Errorln("Error from copy")
      p.Stop()
      return
    }
  }
}

var remoteAddr *string = flag.String("r", "boom", "remote address")

func main() {

    flag.Parse()
    log.SetLevel(log.InfoLevel)

    if *remoteAddr == "boom" {
      panic("Specify proxy server address!")
    }
    NewProxy("localhost:1111", *remoteAddr).Start()
    fmt.Println("Server started.")
    select{}
}