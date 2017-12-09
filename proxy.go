// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package proxy

import (
  "io"
  "net"
  "sync"
  "fmt"

  log "github.com/Sirupsen/logrus"
)

type Proxy struct {
  from string
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
      fmt.Printf("Connectoin from %s\n", connection.LocalAddr().String())
      fmt.Printf("Connectoin to %s\n", connection.RemoteAddr().String())
      if err == nil {
        go p.handle(connection)
      } else {
        p.log.WithField("err", err).Errorln("Error accepting conn")
      }
    }
  }
}

func (p *Proxy) handle(connection net.Conn) {

  defer connection.Close()
  p.log.Debugln("Handling", connection)
  defer p.log.Debugln("Done handling", connection)
  remote, err := net.Dial("tcp", connection.RemoteAddr().String())
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
