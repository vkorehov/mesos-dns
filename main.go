package main

import (
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/resolver"
	"github.com/miekg/dns"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	var resolver resolver.Resolver
	resolver.Config = records.SetConfig()

	log.Println("num of goroutines" + strconv.Itoa(runtime.NumGoroutine()))

	resolver.SetupCon()

	// reload the first time
	resolver.Reload()
	ticker := time.NewTicker(time.Second * time.Duration(resolver.Config.Refresh))
	go func() {
		for _ = range ticker.C {
			log.Println("num of goroutines" + strconv.Itoa(runtime.NumGoroutine()))
			resolver.Reload()
		}
	}()

	// handle for everything in this domain...
	dns.HandleFunc(resolver.Config.Domain+".", panicRecover(resolver.HandleMesos))
	dns.HandleFunc(".", panicRecover(resolver.HandleNonMesos))

	go resolver.Serve("tcp")
	go resolver.Serve("udp")

	wg.Add(1)
	wg.Wait()
}

func panicRecover(f func(w dns.ResponseWriter, r *dns.Msg)) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		defer func() {
			if rec := recover(); rec != nil {
				m := new(dns.Msg)
				m.SetReply(r)
				m.SetRcode(r, 2)
				_ = w.WriteMsg(m)
				log.Println("num of goroutines" + strconv.Itoa(runtime.NumGoroutine()))

				log.Println(rec)
			}
		}()
		f(w, r)
	}
}
