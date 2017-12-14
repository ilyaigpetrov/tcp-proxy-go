package main

import (
  "net"
  "fmt"
  "syscall"
  "flag"
  "sync"

  "github.com/chifflier/nfqueue-go/nfqueue"
)

func run(payload *nfqueue.Payload) int {

  fmt.Println("run")
  handle(payload.Data)
  return nfqueue.NF_ACCEPT

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

  fmt.Println("lock")
  mutex.Lock()
  fmt.Println("write data...")
  wcount, err := remote.Write(data)
  mutex.Unlock()
  if err != nil {
    panic(err)
  }
  if wcount != len(data) {
    panic(fmt.Sprintf("Not all data written: %s/%s", wcount, len(data)))
  }

}

var remoteAddr *string = flag.String("r", "boom", "remote address")

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
