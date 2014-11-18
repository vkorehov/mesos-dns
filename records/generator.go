// package records contains functions to generate resource records from
// mesos master states to serve through a dns server
package records

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// rrs is a type of question names to resource records answers
type rrs map[string][]string

// Slaves is a mapping of id to hostname read in from state.json
type Slaves []struct {
	Id       string `json:"id"`
	Hostname string `json:hostname"`
}

// Resources holds our SRV ports
type Resources struct {
	Ports string `json:"ports"`
}

// Tasks holds mesos task information read in from state.json
type Tasks []struct {
	FrameworkId string `json:"framework_id"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	SlaveId     string `json:"slave_id"`
	State       string `json:"state"`
	Resources   `json:"resources"`
}

// Frameworks holds mesos frameworks information read in from state.json
type Frameworks []struct {
	Tasks `json:"tasks"`
	Name  string `json:"name"`
}

// StateJSON is a representation of mesos master state.json
type StateJSON struct {
	Frameworks `json:"frameworks"`
	Slaves     `json:"slaves"`
	Leader     string `json:"leader"`
}

// RecordGenerator is a tmp mapping of resource records and slaves
// maybe de-dupe
// prob. want to break apart
// refactor me - prob. not needed
type RecordGenerator struct {
	RRs rrs
	Slaves
}

// hostBySlaveId looks up a hostname by slave_id
func (rg *RecordGenerator) hostBySlaveId(slaveId string) (string, error) {
	for i := 0; i < len(rg.Slaves); i++ {
		if rg.Slaves[i].Id == slaveId {
			return rg.Slaves[i].Hostname, nil
		}
	}

	return "", errors.New("not found")
}

// loadFromFile loads fake state.json from a config file
// does not belong here - mv to test
func (rg *RecordGenerator) loadFromFile() (sj StateJSON) {
	b, err := ioutil.ReadFile("test/fake.json")
	if err != nil {
		fmt.Println("missing test data")
	}

	err = json.Unmarshal(b, &sj)
	if err != nil {
		fmt.Println(err)
	}

	return sj
}

// loadFromMaster loads state.json from mesos master
func (rg *RecordGenerator) loadFromMaster(ip string, port string) (sj StateJSON) {

	fmt.Println("reloading using " + ip)

	// tls ?
	url := "http://" + ip + ":" + port + "/master/state.json"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(body, &sj)
	if err != nil {
		fmt.Println(err)
	}

	return sj
}

// stripUID removes the UID from a taskName
func stripUID(taskName string) string {
	return strings.Split(taskName, ".")[0]
}

// leaderIP returns the ip for the mesos master
func leaderIP(leader string) string {
	pair := strings.Split(leader, "@")[1]
	return strings.Split(pair, ":")[0]
}

// loadWrap catches an attempt to load state.json from a mesos master
// attempts can fail from down server or mesos master secondary
// it also reloads from a different master if the master it attempted to
// load from was not the leader
func (rg *RecordGenerator) loadWrap(ip string, port string) (StateJSON, error) {
	var err error
	var sj StateJSON

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("can't connect to mesos")
		}

	}()

	sj = rg.loadFromMaster(ip, port)

	if rip := leaderIP(sj.Leader); rip != ip {
		sj = rg.loadFromMaster(rip, port)
	}

	return sj, err
}

// yankPort grabs the first port in the port field
// this takes a string even though it should take an array
func yankPort(ports string) string {
	rhs := strings.Split(ports, "[")[1]
	lhs := strings.Split(rhs, "]")[0]
	return strings.Split(lhs, "-")[0]
}

// ParseState parses a state.json from a mesos master
// it sets the resource records map for the resolver
// with the following format
//
//  _<tag>.<service>.<framework>._<protocol>..mesos
// it also tries different mesos masters if one is not up
// this will shudown if it can't connect to a mesos master
func (rg *RecordGenerator) ParseState(config Config) {

	var sj StateJSON

	port := strconv.Itoa(config.Port)

	// try each listed mesos master before dying
	for i := 0; i < len(config.Masters); i++ {
		sj, _ = rg.loadWrap(config.Masters[i], port)

		if sj.Leader == "" {
			fmt.Println("no leader - trying next one")

			if len(config.Masters)-1 == i {
				os.Exit(2)
			}

		} else {
			break
		}

	}

	rg.Slaves = sj.Slaves

	rg.RRs = make(rrs)

	f := sj.Frameworks

	// complete crap - refactor me
	for i := 0; i < len(f); i++ {
		fname := f[i].Name

		for x := 0; x < len(f[i].Tasks); x++ {
			task := f[i].Tasks[x]

			host, err := rg.hostBySlaveId(task.SlaveId)
			if err == nil {
				tname := stripUID(task.Name)
				port := yankPort(task.Resources.Ports)

				// hack
				host += ":" + port

				tcp := tname + "._tcp." + fname + ".mesos."
				udp := tname + "._udp." + fname + ".mesos."

				tcpnof := tname + "._tcp" + ".mesos."
				udpnof := tname + "._udp" + ".mesos."

				fmt.Println(tcp + " " + host)
				fmt.Println(udp + " " + host)
				fmt.Println(tcpnof + " " + host)
				fmt.Println(udpnof + " " + host)

				if val, ok := rg.RRs[tcp]; ok {
					rg.RRs[tcp] = append(val, host)
				} else {
					rg.RRs[tcp] = []string{host}
				}

				if val, ok := rg.RRs[udp]; ok {
					rg.RRs[udp] = append(val, host)
				} else {
					rg.RRs[udp] = []string{host}
				}

				if val, ok := rg.RRs[tcpnof]; ok {
					rg.RRs[tcpnof] = append(val, host)
				} else {
					rg.RRs[tcpnof] = []string{host}
				}

				if val, ok := rg.RRs[udpnof]; ok {
					rg.RRs[udpnof] = append(val, host)
				} else {
					rg.RRs[udpnof] = []string{host}
				}

			}
		}
	}

}
