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
	vars        map[string]string
}

// Return a list of UPSInfo with just name and description populated.
func getUPSNames(upss *string, port *int) []*UPSInfo {
	ret := []*UPSInfo{}

	cmd_output := upsdCommand(upss, port, "LIST UPS")

	lines := strings.Split(cmd_output, "\n")

	for _, l := range lines {
		if len(l) > 3 && l[:3] == "UPS" {
			ret = append(ret, getUPSInfoFromListUPSOutput(l))
		}
	}

	return ret
}

// Get all variables for a given UPS. Does not update UPSInfo.
func getUpsVars(upss *string, port *int, u *UPSInfo) (map[string]string, error) {
	ret := make(map[string]string)
	cmd_output := upsdCommand(upss, port, "LIST VAR "+u.name)
	lines := strings.Split(cmd_output, "\n")

	for _, l := range lines {
		if len(l) > 3 && l[:3] == "VAR" {
			fragments := strings.Split(l, " ")
			k := fragments[2]
			v := strings.Join(fragments[3:], " ")
			// Strip quotes
			v = v[1:(len(v) - 1)]
			ret[k] = v
		}
	}

	return ret, nil
}

// Format: UPS <upsname> "Description"
func getUPSInfoFromListUPSOutput(line string) *UPSInfo {
	fragments := strings.Split(line, " ")
	desc := string(strings.Split(line, "\"")[1])
	var ret UPSInfo
	ret.name = fragments[1]
	ret.description = desc
	return &ret
}

// Issue a ommand to upsd and return the output.
func upsdCommand(upss *string, port *int, cmd string) (rep string) {
	addr := *upss + ":" + strconv.Itoa(*port)
	c, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	c.Write([]byte(cmd + "\n"))

	buf := make([]byte, 2048)
	tmp := make([]byte, 256)

	for {
		n, _ := c.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if strings.Contains(string(tmp), "END "+cmd) {
			break
		}
	}

	return string(buf)
}
