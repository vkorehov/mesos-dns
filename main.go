package main

import (
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/resolver"
	"github.com/miekg/dns"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	var resolver resolver.Resolver
	resolver.Config = records.SetConfig()

	// reload the first time
	resolver.Reload()
	ticker := time.NewTicker(time.Second * time.Duration(resolver.Config.Refresh))
	go func() {
		for _ = range ticker.C {
			resolver.Reload()
		}
	}()

	// handle for everything in this domain...
	dns.HandleFunc("mesos.", resolver.HandleMesos)

	go resolver.Serve("tcp")
	go resolver.Serve("udp")

	wg.Add(1)
	wg.Wait()
}
