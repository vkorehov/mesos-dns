mesos-dns
=========

__Tell GO how many cores to use__
  ```
  # ulimit -n 16000             # set file descriptors high (to get more goroutines)
  # GOMAXPROCS=4 ./mesos-dns
  ```

__Benchmarks on GCE Dev Cluster__

__Memory Usage on GCE Dev Cluster__
