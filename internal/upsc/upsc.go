package upsc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type UPSDClient struct {
	host string
	port int
}

func NewUPSDClient(host string, port int) *UPSDClient {
	return &UPSDClient{host: host, port: port}
}

type UPSInfo struct {
	name        string
	description string
	vars        map[string]string
}

func (u *UPSInfo) Name() string {
	return u.name
}

func UpsdCommand(upsd_c *UPSDClient, cmd string) (map[string]string, error) {
	// Get raw output from upsd if we know how to parse it.
	if strings.HasPrefix(cmd, "LIST") || strings.HasPrefix(cmd, "GET") {
		raw, err := rawUpsdCommand(upsd_c, cmd)
		if err != nil {
			return nil, err
		}
		return processUpsdResponse(raw, cmd)
	} else {
		return nil, errors.New("Don't know how to issue upsd command: " + cmd)
	}
}

func processUpsdResponse(response string, cmd string) (map[string]string, error) {
	ret := map[string]string{}
	replines := strings.Split(response, "\n")
	// Drop trailing newline
	if len(replines) > 0 && replines[len(replines)-1] == "" {
		replines = replines[:len(replines)-1]
	}
	if len(replines) == 0 {
		// No response is fine.
		return ret, nil
	}
	if strings.HasPrefix(cmd, "LIST") {
		// rfc9271 4.2.7 "All the LIST commands had fucking better produce a response with a common format."
		// (I'm paraphrasing here)
		if !strings.HasPrefix(replines[0], fmt.Sprintf("BEGIN %v", cmd)) {
			return nil, fmt.Errorf("no BEGIN preamble in %v response", cmd)
		}
		if !strings.HasPrefix(replines[len(replines)-1], fmt.Sprintf("END %v", cmd)) {
			return nil, fmt.Errorf("no END addendum in %v response", cmd)
		}
		if len(replines) == 2 {
			return ret, nil
		}
		for i := 1; i < (len(replines) - 1); i++ {
			k, v, err := getKeyValueFromListLine(replines[i])
			if err != nil {
				return nil, err
			}
			ret[k] = v
		}
		return ret, nil
	}

	if strings.HasPrefix(cmd, "GET") {
		// GET should return a single item, if at all.
		if len(replines) > 1 {
			return nil, fmt.Errorf("multiple response lines from GET command")
		}
		k, v, err := getKeyValueFromListLine(replines[0])
		if err != nil {
			return nil, err
		}
		if k != "" {
			ret[k] = v
		}
		return ret, nil
	}

	return nil, fmt.Errorf("command not implemented: %v", cmd)

}

func getKeyValueFromListLine(line string) (string, string, error) {
	// See rfc9271, sections 4.2.4 and 4.2.7
	// Each line in either a single-line GET response or a multi-line LIST response is of the form:
	// NOUN <requested> <response>
	// <requested> can of course be more then one token, of course. The response is the rest of the line.
	// So here, we figure things out by noun, and shit the bed of we don't 100% know how the noun works.
	if line == "" {
		return "", "", nil
	}
	fragments := strings.Split(line, " ")
	switch fragments[0] {
	case "UPS":
		// UPS upsname "ups description"
		val_raw := strings.Join(fragments[2:], " ")
		return fragments[1], val_raw[1 : len(val_raw)-1], nil
	case "VAR":
		// VAR myups varname "var value"
		val_raw := strings.Join(fragments[3:], " ")
		return fragments[2], val_raw[1 : len(val_raw)-1], nil
	default:
		return "", "", fmt.Errorf("do not know how to interpret UPS response: %v", line)
	}

}

func GetUPSes(upsd_c *UPSDClient) ([]*UPSInfo, error) {
	// Get the list of UPSes from upsd
	upses := []*UPSInfo{}
	upslist, err := UpsdCommand(upsd_c, "LIST UPS")
	if err != nil {
		return nil, err
	}
	for k, v := range upslist {
		log.Printf("Found UPS: %v (%v)\n", k, v)
		upses = append(upses, &UPSInfo{name: k, description: v, vars: make(map[string]string)})
	}
	return upses, nil
}

func GetUpdatedVars(upsd_c *UPSDClient, u *UPSInfo) (map[string]string, error) {
	// Fetch updated vars for this UPS and both update the struct in place and return the new values.
	ret := map[string]string{}
	new_vars, err := UpsdCommand(upsd_c, "LIST VAR "+u.Name())
	if err != nil {
		return nil, err
	}
	for k, v := range new_vars {
		old_v, present := u.vars[k]
		if !present || old_v != v {
			// variable is new or changed
			ret[k] = v
			u.vars[k] = v
		}
	}
	return ret, nil
}

// Issue a raw command to upsd and return the output.
func rawUpsdCommand(upsd_c *UPSDClient, cmd string) (rep string, err error) {
	addr := upsd_c.host + ":" + strconv.Itoa(upsd_c.port)
	c, err := net.Dial("tcp", addr)

	if err != nil {
		return "", err
	}

	defer c.Close()

	end_marker := "\n"
	if strings.HasPrefix(cmd, "LIST") {
		end_marker = "END LIST"
	}

	c.Write([]byte(cmd + "\n"))

	buf := make([]byte, 2048)

	for {
		raw, err := c.Read(buf)
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		rep += string(buf[:raw])
		if strings.Contains(string(buf), end_marker) {
			break
		}
	}

	return rep, nil
}
