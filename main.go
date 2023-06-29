package main

import (
  "fmt"
  "log"
)

func main() {
  fmt.Println("Hello")

  ups := getUPSNames("localhost", 3493)

  for _, u := range ups {
    fmt.Printf("Found UPS: %v (%v)\n", u.name, u.description)
  }

}

func checkErrFatal(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
