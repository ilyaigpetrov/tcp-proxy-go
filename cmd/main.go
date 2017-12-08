package main

import (
  "fmt"
  proxy "../../proxy"
)

func main() {
    proxy.NewProxy("127.0.0.1", "127.0.0.1")
    fmt.Println("hello world")
}
