package main

import (
  "fmt"
  "log"
  "net"
  "strconv"
)

type UPSInfo struct {
  name string
  info map[string]string
}

func getUPSNames(upss string, port int) string {
  addr := upss + ":" + strconv.Itoa(port)
  fmt.Println("Asking " + addr + " for list.")
  c, err := net.Dial("tcp", addr)

  if err != nil {
    log.Fatal(err)
  }

  defer c.Close()

  c.Write([]byte("LIST UPS\n"))

  rep := make([]byte, 1024)

  c.Read(rep)

  return string(rep)
  return ""
}
