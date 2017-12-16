package main

func main() {

  fromTCP, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1111")
  if (err != nil) {
    panic(err)
  }
  listener, err := net.ListenTCP("tcp", p.fromTCP)
  if (err != nil) {
    panic(err)
  }
  connection, err := listener.AcceptTCP()
  if (err != nil) {
    panic(err)
  }
  defer connection.Close()

  buf := make([]byte, 0, 4096) // big buffer
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
      buf = append(buf, tmp[:n]...)
  }

}
