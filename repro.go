// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package main

import (
  "syscall"
  "github.com/chifflier/nfqueue-go/nfqueue"
)

func run(payload *nfqueue.Payload) int {
  // p.handle(payload.Data)
  return nfqueue.NF_ACCEPT
}

func main() {

  q := new(nfqueue.Queue)

  q.SetCallback(run)

  q.Init()

  q.Unbind(syscall.AF_INET)
  q.Bind(syscall.AF_INET)

  q.CreateQueue(1)

  q.Loop()
  q.DestroyQueue()
  q.Close()

}
