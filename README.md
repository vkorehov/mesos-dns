mesos-dns
=========

DNS for service discovery with Mesos. 
Refer to the [initial design document](https://docs.google.com/a/mesosphere.io/document/d/1h-ptANif4RZNWKTAJXsG0s4ZjfpY7GLrY2zRwmrIBAc/edit?usp=sharing) for details. 

__INSTALL GO__
  ```shell
  sudo apt-get install git-core
  wget https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
  tar xzf go*
  sudo mv go /usr/local/.
  # puts this into ~/.profile
  export PATH=$PATH:/usr/local/go/bin
  export GOROOT=/usr/local/go
  export PATH=$PATH:$GOROOT/bin
  export GOPATH=$HOME/go
  ```
 
__INSTALL MESOS-DNS__

  ```shell
  go get github.com/miekg/dns
  git clone git@github.com:mesosphere/mesos-dns.git
  ```
__BUILD & CONFIGURE MESOS-DNS__
  ```
  cd mesos-dns
  go build -o mesos-dns main.go
  cp config.json.sample config.json 
  (adjust values as you see fit)
  ```

__RUN__
  ```
  // root only needed if you are using port 53 (recommended)
  sudo ./mesos-dns
  ```

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

__Benchmarking__

  https://github.com/jedisct1/dnsblast

  osx:
    sudo sysctl -w kern.maxfiles=60000
    sudo sysctl -w kern.maxfilesperproc=60000
    sudo sysctl -w kern.ipc.somaxconn=60000
    ulimit -S -n 60000

  linux:

    edit /etc/sysctl.conf 
      ```
        fs.file-max = 65536
      ```

    edit /etc/security/limits.conf
      ```
      * soft nproc 65535
      * hard nproc 65535
      * soft nofile 65535
      * hard nofile 65535
      ```

      ulimit -a should show the correct limits

      ```
        cd tools/main.go
        go run main.go
      ```
       
__WARNING__

* no test coverage at the moment
