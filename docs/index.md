---
title: A DNS server for service discovery in Mesos clusters
---

<div class="jumbotron text-center">
  <h1>Mesos-DNS</h1>
  <p class="lead">
    A DNS server for service discovery in Apache Mesos clusters
  </p>
  <p>
    <a href="http://github.com/mesosphere/mesos-dns"
        class="btn btn-lg btn-primary">
      Mesos-DNS repository (v0.1.0 - alpha)
    </a>
  </p>
<!--  <a class="btn btn-link"
      href="http://downloads.mesosphere.com/mesos-dns/v0.1.0/mesos-dns-0.1.0.tgz.sha256">
    v0.1.0 SHA-256 Checksum
  </a> &middot;
  <a class="btn btn-link"
      href="https://github.com/mesosphere/mesos-dns/releases/tag/v0.1.0">
    v0.1.0 Release Notes
  </a>
-->
</div>

## Overview

[Mesos-DNS](http://github.com/mesosphere/mesos-dns) supports service discovery in [Apache Mesos](http://mesos.apache.org/) clusters. It allows applications running in the cluster to find each other through the domain name system ([DNS](http://en.wikipedia.org/wiki/Domain_Name_System)). Mesos-DNS is designed to be a minimal, stateless service that is easy to deploy and maintain.

The figure below depicts how Mesos-DNS works:

<p class="text-center">
  <img src="{{ site.baseurl}}/img/architecture.png" width="610" height="320" alt="">
</p>

Mesos-DNS periodically contacts the Mesos master(s), retrieves the state of all running tasks from all running frameworks, and generates DNS records for these tasks (A and SRV records). As tasks start, finish, fail, or restart on the Mesos cluster, Mesos-DNS updates the DNS records to reflect the latest state. Tasks running on Mesos slaves can discover the IP addresses and ports of other tasks they depend upon by issuing DNS lookup requests. Mesos-DNS replies directly DNS requests for tasks launched by Mesos. For DNS requests for other hostnames or services, Mesos-DNS uses an external nameserver to derive replies.

Mesos-DNS is stateless. On restart after a failure, it retrieves the latest state from the Mesos master(s) and serves DNS requests without further coordination. It can be easily replicated to improve availability or to load balance DNS requests in clusters with large numbers of slaves. 

The current **alpha** version of Mesos-DNS (v0.1.0) has been tested with Mesos version v0.21.0. It has no dependencies to any frameworks. 

