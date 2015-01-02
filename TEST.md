Testing Mesos-DNS
=========

See the README file for basic installation/compilation instructions. 
Make sure that the port used by the DNS server (typically 53) is not blocked by a firewall rule. 

__Manual Tests__
``` 
 sudo apt-get install dnsutils
* test out A records:
  dig @127.0.0.1 -p 53 testing._tcp.marathon-0.7.5.mesos A
  dig @127.0.0.1 -p 53 "testing._tcp.*.mesos" A
  dig @127.0.0.1 -p 53 google.com A

* test out AAAA records:
  dig @127.0.0.1 -p 53 testing._tcp.marathon-0.7.5.mesos AAAA
  dig @127.0.0.1 -p 53 "testing._tcp.*.mesos" AAAA
  dig @127.0.0.1 -p 53 google.com AAAA

* test out SRV records:
  dig @127.0.0.1 -p 8053 testing._tcp.marathon-0.7.5.mesos SRV
  dig @127.0.0.1 -p 8053 "testing._tcp.*.mesos" SRV

* test out ANY records:
  dig @127.0.0.1 -p 8053 bob._tcp.marathon-0.7.5.mesos ANY
  dig @127.0.0.1 -p 8053 "bob._tcp.*.mesos" ANY
``` 

__Testing with [ResPerf](http://linux.die.net/man/1/resperf)__
``` 
wget ftp://ftp.nominum.com/pub/nominum/dnsperf/2.0.0.0/dnsperf-2.0.0.0-1-rhel-6-x86_64.tar.gz
tar xzf dnsperf-2.0.0.0-1-rhel-6-x86_64.tar.gz
sudo apt-get install alien
alien -i dnsperf-2.0.0.0-1.el6.x86_64.rpm 
export PATH=$PATH:/usr/local/nom/bin
wget ftp://ftp.nominum.com/pub/nominum/dnsperf/data/queryfile-example-current.gz
tar xzf queryfile-example-current.gz

# attempt to issue 100 QPS for a rump-up of 10 seconds
resperf -s10.117.207.42 -d queryfile-example-current  -m 100 -r 10
``` 

__other__

* sample.config.json w/full examples
  - dns left out to rely on host system

* test task in test/testing.sh
  cd /home/jclouds && ./testing.sh


__Performance__

Mesos-dns will comfortably scale to 8.5K q/s for internal queries
.
If you wish to leverage more cores please adjust your maxfiles limits and use the GOMAXPROCS=N cores environment variable.

Example:
```
  GOMAXPROCS=8 ./mesos-dns
```

A note on external queries. Many open recursors such as google's 8.8.8.8 will throttle your connection to a much lower number than what the server can actually handle.

If you are experiencing dns request timeouts first thing to check is if it's internal or external requests. If external you might try using a different recursor or set of recursors:

http://public-dns.tk/nameservers

You may also choose to install dnsmasq which can cache external queries.

If you find yourself adjusting gomaxprocs you'll probably want to adjust the maxfiles limits on your operating system as well:

on osx:
  ```
    sudo sysctl -w kern.maxfiles=60000
    sudo sysctl -w kern.maxfilesperproc=60000
    sudo sysctl -w kern.ipc.somaxconn=60000
    ulimit -S -n 60000
  ```
 
on linux:
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

ulimit -a should show the correct limits, if not go ahead and adjust the ulimit in the shell that mesos-dns runs in via:

```
  ulimit -n 60000
```

__Memory Usage__

Go manages memory differently than what you might be used to. It will reserve a large chunk right off the bat (virt) but your (res) is much closer to reality of what is in use. Having said that, res is not even as precise as you might be used to. To get the true values you'll want to look at pprof - in particular the --inuse_space vs. --alloc_space . Allocated will probably be closer to RES than the actual in-use. The kernel is madvised for the difference.

Also - each connection has it's own go-routine whose stack does not get released. It is re-used. This might be fixed in go 1.4.

You can expect under load to see ~ >200M for RES.

__TESTING WITH MESOSAURUS__
 ```
git clone https://github.com/mesosphere/mesosaurus.git
# if scala is not installed
wget http://www.scala-lang.org/files/archive/scala-2.9.3.tgz && tar -xvf  scala-2.9.3.tgz && cd scala-2.9.3 && export PATH=`pwd`/bin:$PATH && export SCALA_HOME=`pwd`
wget https://dl.bintray.com/sbt/native-packages/sbt/0.13.7/sbt-0.13.7.tgz
tar xvf sbt-0.13.7.tgz 
export PATH=`pwd`/sbt/bin:$PATH
# build mesosaurus
cd mesosaurus/task
make
cd ..
sbt compile
# launches 10 tasks of 1000msec duration each
sbt "run -tasks 10 -duration 1000 -arrival 200 -master 10.90.16.131:5050"
 ```

__Unit Testing__
```
go test -v ./...
```

__Scripted Test__
```
#!/bin/sh
echo Running resperf on testfile $1, max QPS $2, duration $3

# run resperf
/usr/local/nom/bin/resperf -s172.16.0.13 -p31054 -d /usr/local/nom/bin/$1  -m $2 -r $3 > resperf.out

# check output
if grep -q "Maximum throughput:   0.000000 qps" resperf.out; then
        echo ERROR, Unable to communicate with mesos-dns server
        /bin/mail -s "Mesos-dns error ($1, $(date))" christos@mesosphere.io < resperf.out
else
        echo Success
fi

```


