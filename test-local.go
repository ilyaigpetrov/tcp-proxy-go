package main

import (
  "fmt"
  "net"
)

func isLocal(targetIp string) bool {

  netInterfaceAddresses, err := net.InterfaceAddrs()
  if err != nil {
    panic(err)
  }
  for _, netInterfaceAddress := range netInterfaceAddresses {

    networkIp, ok := netInterfaceAddress.(*net.IPNet)

    if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {
      ip := networkIp.IP.String()
      if ip == targetIp {
        return true
      }
    }
  }
  return false

}

func main() {

  fmt.Println(isLocal("0.0.0.0"))
  fmt.Println(isLocal("127.0.0.1"))
  fmt.Println(isLocal("10.0.0.2"))

}
