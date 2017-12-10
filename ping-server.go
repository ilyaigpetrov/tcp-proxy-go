// https://gist.github.com/ericflo/7dcf4179c315d8bd714c
package main

import (
  "net"
  "flag"
  "fmt"
)

var port *string = flag.String("p", "1234", "port")

func main() {

    flag.Parse();
    listener, err := net.Listen("tcp", ":" + *port)
    if err != nil {
      panic(err)
    }
    fmt.Println("Server started.")
    for {
      connection, err := listener.Accept()
      if err != nil {
        panic(err)
      }
      defer connection.Close()
      connection.Write([]byte("PING"))
    }

}
