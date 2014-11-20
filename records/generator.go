// package records contains functions to generate resource records from
// mesos master states to serve through a dns server
package records

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// rrs is a type of question names to resource records answers
type rrs map[string][]string

type slave struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
}

// Slaves is a mapping of id to hostname read in from state.json
type Slaves []slave

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
	As   rrs
	SRVs rrs
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

// findMaster tries each master and looks for the leader
// if no leader responds it errors
func (rg *RecordGenerator) findMaster(masters []string, port string) (StateJSON, error) {
	var sj StateJSON

	// try each listed mesos master before dying
	for i := 0; i < len(masters); i++ {
		sj, _ = rg.loadWrap(masters[i], port)

		if sj.Leader == "" {
			log.Println("not a leader - trying next one")

			if len(masters)-1 == i {
				return sj, errors.New("no master")
			}

		} else {
			return sj, nil
		}

	}

	return sj, nil
}

// ParseState parses a state.json from a mesos master
// it sets the resource records map for the resolver
// with the following format
//
//  _<tag>.<service>.<framework>._<protocol>..mesos
// it also tries different mesos masters if one is not up
// this will shudown if it can't connect to a mesos master
func (rg *RecordGenerator) ParseState(config Config) {

	port := strconv.Itoa(config.Port)

	// try each listed mesos master before dying
	sj, err := rg.findMaster(config.Masters, port)
	if err != nil {
		log.Println("no master")
		return
	}

	rg.InsertState(sj, config.Domain)
}

// InsertState transforms a StateJSON into RecordGenerator RRs
func (rg *RecordGenerator) InsertState(sj StateJSON, domain string) error {
	rg.Slaves = sj.Slaves

	rg.SRVs = make(rrs)
	rg.As = make(rrs)

	f := sj.Frameworks

	// complete crap - refactor me
	for i := 0; i < len(f); i++ {
		fname := f[i].Name

		for x := 0; x < len(f[i].Tasks); x++ {
			task := f[i].Tasks[x]

			host, err := rg.hostBySlaveId(task.SlaveId)
			if err == nil {
				tname := stripUID(task.Name)
				sport := yankPort(task.Resources.Ports)

				// hack
				host += ":" + sport
				tail := fname + "." + domain + "."

				tcp := tname + "._tcp." + tail
				udp := tname + "._udp." + tail

				rg.insertRR(tcp, host, "SRV")
				rg.insertRR(udp, host, "SRV")

				arec := tname + "." + tail
				arecnof := tname + "." + domain + "."
				rg.insertRR(arec, host, "A")
				rg.insertRR(arecnof, host, "A")

			}
		}
	}

	return nil
}

// insertRR inserts host to name's map
// refactor me
func (rg *RecordGenerator) insertRR(name string, host string, rtype string) {
	if rtype == "A" {

		if val, ok := rg.As[name]; ok {
			rg.As[name] = append(val, host)
		} else {
			rg.As[name] = []string{host}
		}

	} else {

		if val, ok := rg.SRVs[name]; ok {
			rg.SRVs[name] = append(val, host)
		} else {
			rg.SRVs[name] = []string{host}
		}
	}

}
