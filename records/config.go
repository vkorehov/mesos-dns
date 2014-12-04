package records

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
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

	//  Domain: name of the domain used (default "mesos", ie .mesos domain)
	Domain string

	// DNS server: IP address of the DNS server for forwarded accesses
	DNS string

	// Debug: turn on verbose logging
	Debug bool
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
		c.DNS = GetLocalDNS()
	}

	return c
}

// localAddies returns an array of local ipv4 addresses
func localAddies() []string {
	addies, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
	}

	bad := []string{}

	for i := 0; i < len(addies); i++ {
		ip, _, err := net.ParseCIDR(addies[i].String())
		if err != nil {
			log.Println(err)
		}
		t4 := ip.To4()
		if t4 != nil {
			bad = append(bad, t4.String())
		}
	}

	for i := 0; i < len(bad); i++ {
		log.Println(bad[i])
	}

	return bad
}

// nonLocalAddies only returns non-local ns entries
func nonLocalAddies(cservers []string) []string {
	bad := localAddies()

	good := []string{}

	for i := 0; i < len(cservers); i++ {
		local := false
		for x := 0; x < len(bad); x++ {
			if cservers[i] == bad[x] {
				local = true
			}
		}

		if !local {
			good = append(good, cservers[i])
		}

	}

	return good
}

// GetLocalDNS returns the first nameserver in /etc/resolv.conf
// used for out of mesos domain queries
func GetLocalDNS() string {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	non := nonLocalAddies(conf.Servers)

	// for now choose the first non-local
	return non[0]
}
