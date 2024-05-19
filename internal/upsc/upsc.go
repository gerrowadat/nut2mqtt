package upsc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	control "github.com/gerrowadat/nut2mqtt/internal/control"
	"github.com/gerrowadat/nut2mqtt/internal/metrics"
)

type UPSDClientIf interface {
	Request(cmd string) (string, error)
	Host() string
	Port() int
}

type UPSDClient struct {
	host string
	port int
}

func NewUPSDClient(host string, port int) *UPSDClient {
	return &UPSDClient{host: host, port: port}
}

func (upsd_c *UPSDClient) Request(cmd string) (string, error) {
	return rawUpsdCommand(upsd_c, cmd)
}

func (upsd_c *UPSDClient) Host() string {
	return upsd_c.host
}

func (upsd_c *UPSDClient) Port() int {
	return upsd_c.port
}

type UPSHosts struct {
	Hosts []*UPSDClient
}

func NewUPSHosts(hosts_flag string, default_port int) *UPSHosts {
	ret := &UPSHosts{}
	hosts := []*UPSDClient{}
	for _, host := range strings.Split(hosts_flag, ",") {
		// If the host has a port, use it. Otherwise, use the default.
		host_fragments := strings.Split(host, ":")
		if len(host_fragments) == 1 {
			hosts = append(hosts, NewUPSDClient(host, 3493))
		} else if len(host_fragments) == 2 {
			port, err := strconv.Atoi(host_fragments[1])
			if err != nil {
				log.Fatalf("Error parsing port number from %v: %v", host, err)
			}
			hosts = append(hosts, NewUPSDClient(host_fragments[0], port))
		} else {
			log.Fatalf("Error parsing host '%v' from flag", host)
		}
	}
	ret.Hosts = hosts
	return ret
}

func (ups_hosts *UPSHosts) UPSInfoProducer(c *control.Controller, mr *metrics.MetricRegistry, wg *sync.WaitGroup, next chan *channels.UPSInfo, poll_interval time.Duration) {
	// Check the list of UPses and emit each one on the channel for checking.
	defer wg.Done()
	for _, upsd_c := range ups_hosts.Hosts {
		log.Printf("Watching for UPSes on %v:%v\n", upsd_c.Host(), upsd_c.Port())
	}
	for {
		for _, upsd_c := range ups_hosts.Hosts {
			upses, err := GetUPSes(upsd_c)
			if err != nil {
				c.Shutdown("Error getting UPSes: %v", err)
			}
			for _, u := range upses {
				mr.Metrics().UPSScrapesCount.Inc()
				GetUpdatedVars(upsd_c, u)
				next <- u
			}
		}
		time.Sleep(poll_interval * time.Second)
	}
}

func UpsdCommand(upsd_c UPSDClientIf, cmd string) (map[string]string, error) {
	// Get raw output from upsd if we know how to parse it.
	if strings.HasPrefix(cmd, "LIST") || strings.HasPrefix(cmd, "GET") {
		//raw, err := rawUpsdCommand(upsd_c, cmd)
		raw, err := upsd_c.Request(cmd)
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

func GetUPSes(upsd_c UPSDClientIf) ([]*channels.UPSInfo, error) {
	// Get the list of UPSes from upsd
	upses := []*channels.UPSInfo{}
	upslist, err := UpsdCommand(upsd_c, "LIST UPS")
	if err != nil {
		return nil, err
	}
	for k, v := range upslist {
		upses = append(upses, &channels.UPSInfo{Name: k, Description: v, Host: upsd_c.Host(), Vars: make(map[string]string)})
	}
	return upses, nil
}

func GetUpdatedVars(upsd_c UPSDClientIf, u *channels.UPSInfo) (map[string]string, error) {
	// Fetch updated vars for this UPS and both update the struct in place and return the new values.
	ret := map[string]string{}
	new_vars, err := UpsdCommand(upsd_c, "LIST VAR "+u.Name)
	if err != nil {
		return nil, err
	}
	for k, v := range new_vars {
		old_v, present := u.Vars[k]
		if !present || old_v != v {
			// variable is new or changed
			ret[k] = v
			u.Vars[k] = v
		}
	}
	for k := range u.Vars {
		_, present := new_vars[k]
		if !present {
			// variable is no longer present
			ret[k] = ""
			delete(u.Vars, k)
		}
	}
	return ret, nil
}

func GetVarDiff(old *channels.UPSInfo, new *channels.UPSInfo) map[string]string {
	// Compare two UPSInfo structs and return a map of new or changed variables.
	ret := map[string]string{}
	for k, v := range new.Vars {
		old_v, present := old.Vars[k]
		if !present || old_v != v {
			ret[k] = v
		}
	}
	return ret
}

// Issue a raw command to upsd and return the output.
func rawUpsdCommand(upsd_c UPSDClientIf, cmd string) (rep string, err error) {
	addr := upsd_c.Host() + ":" + strconv.Itoa(upsd_c.Port())
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
