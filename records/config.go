package records

import (
	"encoding/json"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/miekg/dns"
	"io/ioutil"
	"net"
	"os"
)

// Config holds mesos dns configuration
type Config struct {

	// Mesos master(s): a list of IP:port/zk pairs for one or more Mesos masters
	Masters []string

	// Refresh frequency: the frequency in seconds of regenerating records (default 60)
	RefreshSeconds int

	// TTL: the TTL value used for SRV and A records (default 60)
	TTL int

	// Resolver port: port used to listen for slave requests (default 53)
	Port int

	//  Domain: name of the domain used (default "mesos", ie .mesos domain)
	Domain string

	// DNS server: IP address of the DNS server for forwarded accesses
	DNS []string
}

// SetConfig instantiates a Config struct read in from config.json
func SetConfig() (c Config) {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		logging.Error.Println("missing config")
	}

	err = json.Unmarshal(b, &c)
	if err != nil {
		logging.Error.Println(err)
	}

	if len(c.DNS) == 0 {
		c.DNS = GetLocalDNS()
	}

	return c
}

// localAddies returns an array of local ipv4 addresses
func localAddies() []string {
	addies, err := net.InterfaceAddrs()
	if err != nil {
		logging.Error.Println(err)
	}

	bad := []string{}

	for i := 0; i < len(addies); i++ {
		ip, _, err := net.ParseCIDR(addies[i].String())
		if err != nil {
			logging.Error.Println(err)
		}
		t4 := ip.To4()
		if t4 != nil {
			bad = append(bad, t4.String())
		}
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
func GetLocalDNS() []string {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		logging.Error.Println(err)
		os.Exit(2)
	}

	non := nonLocalAddies(conf.Servers)

	// for now choose the first non-local
	return non[0]
}
