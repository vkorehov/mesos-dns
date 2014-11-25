mesos-dns
=========

DNS for service discovery with Mesos. 
Refer to the [initial design document](https://docs.google.com/a/mesosphere.io/document/d/1h-ptANif4RZNWKTAJXsG0s4ZjfpY7GLrY2zRwmrIBAc/edit?usp=sharing) for details. 

__Install go__
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
 
__Build mesos-dns__

  ```shell
  go get github.com/miekg/dns
  git clone git@github.com:mesosphere/mesos-dns.git
  cd mesos-dns
  go build -o mesos-dns main.go
  ```

__Configure mesos-dns server__
  ```shell
  cp config.json.sample config.json 
  # edit config.json
  # "masters" --> list of IP addresses for mesos masters
  # "port" --> REST API port for mesos masters
  # "resolver" --> port for mesos-dns (53 is strongly recommended)
  # "domain" --> the DNS domain for the mesos cluster
  ```
__Configure mesos slaves__
  ```
  # edit /etc/resolv.conf
  # add "nameserver <ip address of mesos-dns>" at the very beginning
  ```

__Run__
  ```shell
  // root only needed if you are using port 53 (recommended)
  sudo ./mesos-dns
  ```

__WARNINGS__

* no test coverage at the moment
