Testing Mesos-DNS
=========

See the README file for basic installation/compilation instructions. 
Make sure that the port used by the DNS server (typically 53) is not blocked by a firewall rule. 

__Manual Tests__
``` 
 sudo apt-get install dnsutils
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
``` 

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
