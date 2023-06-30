package main

import (
	"log"
	"net"
	"strconv"
	"strings"
)

type UPSInfo struct {
	name        string
	description string
	info        map[string]string
}

// Return a list of UPSInfo with just name and description populated.
func getUPSNames(upss *string, port *int) []UPSInfo {
	addr := *upss + ":" + strconv.Itoa(*port)
	c, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	c.Write([]byte("LIST UPS\n"))

	buf := make([]byte, 2048)
	tmp := make([]byte, 256)

	ret := []UPSInfo{}

	for {
		n, _ := c.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if strings.Contains(string(tmp), "END LIST UPS") {
			break
		}
	}

	lines := strings.Split(string(buf), "\n")

	for _, l := range lines {
		if len(l) > 3 && l[:3] == "UPS" {
			ret = append(ret, getUPSInfoFromListUPSOutput(l))
		}
	}

	return ret
}

// Format: UPS <upsname> "Description"
func getUPSInfoFromListUPSOutput(line string) UPSInfo {
	fragments := strings.Split(line, " ")
	desc := string(strings.Split(line, "\"")[1])
	return UPSInfo{name: fragments[1], description: desc}
}
