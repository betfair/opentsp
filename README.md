
Project Goals
-------------

   * High compatibility with tcollector

    * Use any language to develop collection plugins that transform metric data into time series
    * Execute the plugins using a low-overhead, contract-checking, supervising forwarder agent
    * Enrich time series on the fly by name translation (adjust metric name, set extra tags)
    * Forward time series to OpenTSDB for historical plotting
    * ... and to custom real-time analysers, export gateways, etc.


  * Forwarding throughput of hundreds of thousands of data points per second
  * User-defined routing topology enabling interoperability with external systems (data integration by dual write)
  * Centralised, programmatic access to real-time feed observing data points with latency on the scale of milliseconds
  * Centralised, programmatic configuration of selected aspects of the pipeline, enabling automation and self-service
  * Self-contained: built in message queueing and configuration management means TSP is free from dependency loops and
    thus suitable for monitoring infrastructure services:

    * RabbitMQ/Kafka
    * Chef/Puppet
    * OpenTSDB
    * ... and so on


  * Pervasive collection:

    * collect-snmp &mdash; high-performance SNMP poller
    * collect-statse &mdash; derive time series from event-level metrics through continuous aggregation
    * collect-netscaler &mdash; high-performance Citrix NetScaler poller (based on Nitro API)
    * collect-f5 &mdash; high-performance F5 BIGIP poller (based on iControl API)
    * ... in addition to the plugins shipped with tcollector


Getting Started
---------------

Follow the quick_start guide.


Next steps
----------

  * Store metrics in database (OpenTSDB)
  * Collect operating system metrics
  * Collect event-level metrics (Statse)
  * Collect network metrics (SNMP)
  * Enrich time series (set extra tags)
  * Run time series health analyser


Documentation
-------------

component      |
---------------|----
tsp-forwarder  | man
tsp-controller | man
tsp-aggregator | man
tsp-poller     | man
collect-statse | man | spec


Issues
------

Use Github Issues to report issues or to get help.


Authors
-------

Jacek Masiulaniec developed TSP based on 3-year experience of running OpenTSDB at Betfair. TSP's design is strongly influenced by the excellent tcollector package, developed by Mark Smith, Dave Barr, and Beno&icirc;t Sigoure.
