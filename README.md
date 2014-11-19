mesos-dns
=========

DNS and service discovery for Mesos.
Initial design document at https://docs.google.com/a/mesosphere.io/document/d/1h-ptANif4RZNWKTAJXsG0s4ZjfpY7GLrY2zRwmrIBAc/edit?usp=sharing. 

__INSTALL__

  go get github.com/miekg/dns
  cp config.json.sample config.json (adjust values as you see fit)

__INSTALL GO__
  wget https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
  tar xzf go*
  sudo mv go /usr/local/.
  export PATH=$PATH:/usr/local/go/bin
  export GOROOT=/usr/local/go
  export PATH=$PATH:$GOROOT/bin
  export GOPATH=$HOME/go

__RUN__

  cp config.json.sample to config.json
  edit to your desires
  ensure that if you are running 53 you need to be root

  ./mesos-dns

__TEST__

  go test -v ./...

  (no tests atm)

__Manual Tests__

* test out A records:

  dig @127.0.0.1 -p 8053 testing._tcp.marathon-0.7.5.mesos A

  dig @127.0.0.1 -p 8053 "testing._tcp.*.mesos" A

  dig @127.0.0.1 -p 8053 google.com A
  (don't support)

* test out AAAA records:

  dig @127.0.0.1 -p 8053 testing._tcp.marathon-0.7.5.mesos AAAA

  dig @127.0.0.1 -p 8053 "testing._tcp.*.mesos" AAAA


* test out SRV records:

  dig @127.0.0.1 -p 8053 testing._tcp.marathon-0.7.5.mesos SRV

  dig @127.0.0.1 -p 8053 "testing._tcp.*.mesos" SRV

* test out ANY records:

  dig @127.0.0.1 -p 8053 bob._tcp.marathon-0.7.5.mesos ANY
  dig @127.0.0.1 -p 8053 "bob._tcp.*.mesos" ANY

__other__

* sample.config.json w/full examples
  - dns left out to rely on host system

* test task in test/testing.sh
  cd /home/jclouds && ./testing.sh

__TODO__

* test GCE currently responds to dig/nslookup but host system currently
  is not using resource records served up by the server

* general benchmarking


__WARNING__

* no test coverage at the moment
