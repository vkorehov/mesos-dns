// package resolver contains functions to handle resolving .mesos
// domains
package resolver

import (
	"fmt"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/miekg/dns"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// resolveOut queries other nameserver
// maybe don't need this - not in use
func (res *Resolver) resolveOut(dom string) {

	nameserver := res.Config.DNS + ":53"

	qt := dns.TypeA
	qc := uint16(dns.ClassINET)

	c := new(dns.Client)
	c.Net = "udp"

	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{dns.Fqdn(dom), qt, qc}

	in, rtt, err := c.Exchange(m, nameserver)
	fmt.Println(rtt)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(in)
}

// cleanWild strips any wildcards out thus mapping cleanly to the
// original serviceName
func cleanWild(dom string) string {
	if strings.Contains(dom, ".*") {
		return strings.Replace(dom, ".*", "", -1)
	} else {
		return dom
	}
}

// splitDomain splits dom into host and port pair
func (res *Resolver) splitDomain(dom string) (host string, port int) {
	s := strings.Split(dom, ":")
	host = s[0]

	// As won't have ports
	if len(s) == 1 {
		return host, 0
	} else {
		port, _ = strconv.Atoi(s[1])
		return host, port
	}
}

// formatSRV returns the SRV resource record for target
func (res *Resolver) formatSRV(name string, target string) *dns.SRV {
	ttl := uint32(res.Config.TTL)

	h, p := res.splitDomain(target)

	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Priority: 0,
		Weight:   0,
		Port:     uint16(p),
		Target:   h + ".",
	}
}

// formatA returns the A resource record for target
func (res *Resolver) formatA(dom string, target string) (*dns.A, error) {
	ttl := uint32(res.Config.TTL)

	h, _ := res.splitDomain(target)

	ip, err := net.ResolveIPAddr("ip4", h)
	if err != nil {
		return nil, err
	} else {

		a := ip.IP

		return &dns.A{
			Hdr: dns.RR_Header{
				Name:   dom,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    ttl},
			A: a.To4(),
		}, nil
	}
}

// shuffleAnswers reorders answers for very basic load balancing
func shuffleAnswers(answers []dns.RR) []dns.RR {
	rand.Seed(time.Now().UTC().UnixNano())

	n := len(answers)
	for i := 0; i < n; i++ {
		r := i + rand.Intn(n-i)
		answers[r], answers[i] = answers[i], answers[r]
	}

	return answers
}

// HandleMesos is a resolver request handler that responds to a resource
// question with resource answer(s)
// it can handle {A, SRV, ANY}
func (res *Resolver) HandleMesos(w dns.ResponseWriter, r *dns.Msg) {
	dom := cleanWild(r.Question[0].Name)
	qType := r.Question[0].Qtype

	m := new(dns.Msg)
	m.Authoritative = true
	m.RecursionAvailable = true
	m.SetReply(r)

	if qType == dns.TypeSRV {

		for i := 0; i < len(res.rs.SRVs[dom]); i++ {
			rr := res.formatSRV(r.Question[0].Name, res.rs.SRVs[dom][i])
			m.Answer = append(m.Answer, rr)
		}

	} else if qType == dns.TypeA {

		for i := 0; i < len(res.rs.As[dom]); i++ {
			rr, err := res.formatA(dom, res.rs.As[dom][i])
			if err != nil {
				fmt.Println(err)
			} else {
				m.Answer = append(m.Answer, rr)
			}

		}

	} else if qType == dns.TypeANY {

		// refactor me
		for i := 0; i < len(res.rs.As[dom]); i++ {
			a, err := res.formatA(r.Question[0].Name, res.rs.As[dom][i])
			if err != nil {
				fmt.Println(err)
			} else {
				m.Answer = append(m.Answer, a)
			}
		}

		for i := 0; i < len(res.rs.SRVs[dom]); i++ {
			srv := res.formatSRV(dom, res.rs.SRVs[dom][i])
			m.Answer = append(m.Answer, srv)
		}

	}

	// shuffle answers
	m.Answer = shuffleAnswers(m.Answer)

	err := w.WriteMsg(m)
	if err != nil {
		fmt.Println(err)
	}
}

// Serve starts a dns server for net protocol
func (res *Resolver) Serve(net string) {

	server := &dns.Server{
		Addr:       ":" + strconv.Itoa(res.Config.Resolver),
		Net:        net,
		TsigSecret: nil}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Failed to setup "+net+" server: %s\n", err.Error())
	}
}

// Resolver holds configuration information and the resource records
// refactor me
type Resolver struct {
	rs     records.RecordGenerator
	Config records.Config
}

// Reload triggers a new refresh from mesos master
func (res *Resolver) Reload() {
	res.rs = records.RecordGenerator{}
	res.rs.ParseState(res.Config)
}