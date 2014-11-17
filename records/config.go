package records

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"os"
)

// Config holds mesos dns configuration
type Config struct {

	// Mesos master(s): a list of IP addresses for one or more Mesos masters
	Masters []string

	// Master port: the port number for accessing master state (default 5050)
	Port int

	// Refresh frequency: the frequency in seconds of regenerating records (default 60)
	Refresh int

	// TTL: the TTL value used for SRV and A records (default 60)
	TTL int

	// Resolver port: port used to listen for slave requests (default 53)
	Resolver int

	// DNS server: IP address of the DNS server for forwarded accesses
	DNS string
}

// SetConfig instantiates a Config struct read in from config.json
func SetConfig() (c Config) {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("missing config")
	}

	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Println(err)
	}

	if c.DNS == "" {
		conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		c.DNS = conf.Servers[0]
	}

	return c
}
