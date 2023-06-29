package main

import (
  "fmt"
)

func main() {
  fmt.Println("Hello")

  ups := getUPSNames("localhost", 3493)

  fmt.Printf(ups)

}
