# SysEng Prometheus Exporter

---

To simplify things, [Docker version>=v17.05.0-ce-rc1][1], with multi-stage image building support,
is required to run the example.

---

## Running Example

One must to have [Docker](https://docker.com) and [docker-compose](https://docs.docker.com/compose/)
installed on the development machine.

Run

~~~
$ docker-compose -f example/docker-compose.yml -p syseng-challenge up
~~~

The command above builds a Docker image with the syseng-exporter, using provided `Dockerfile`;
downloads supporting images with syseng-challenge service and Prometheus; and runs everything.

Prometheus is running at http://localhost:9090, open the address in web browser.

The following metrics are exported from the syseng-challenge service:

- `syseng_http_requests_total{code="NNN"}` - shows how often each HTTP status code has been served
  during the lifetime of the syseng-challenge binary. It's equivalent to `requestCounters.NNN` stat,
  provided by syseng-challenge.
- `syseng_http_request_duration_seconds_count` - how many requests have been served in total during
  the lifetime of the syseng-challenge binary. Equivalent to `duration.count` stat.
- `syseng_http_request_duration_seconds_sum` - how much total time those requests have taken.
  Equivalent to `duration.sum` stat.
- `syseng_up` - indicates whether the last scrap succeeded.

syseng-challenge exposes `requestRates` and `duration.average` stats with agregated over time values.
They can be calculated using total / summary values and Prometheus query functions, so it's unnessecary
to export them.

For example, `rate(syseng_http_requests_total{code="200"}[1m])` calculates the per-second rate (QPS)
for requests with status code 200 over the last 1 minute.

See "[Drop less useful statistics](https://prometheus.io/docs/instrumenting/writing_exporters/#drop-less-useful-statistics)"
section from Prometheus own documentation on writing exporters.

## Bonus Questions

> 1. What are good ways of deploying hundreds of instances of our simulated service?
     How would you deploy your exporter? And how would you configure Prometheus to monitor them all?

Docker has [docker-stack](https://docs.docker.com/engine/reference/commandline/stack/) which might be helpful
for scaling. The exporter could be run as a sidecar container, so each syseng-challege instance is linked
to the exporter instance in a 1-to-1 way.

Solutions like [Nomad](https://www.nomadproject.io) or [Kubernetes](https://kubernetes.io) have concepts
of groupping for logicaly tied services (syseng-challege and the exporter): the former provides groups,
while the latter has Pods.

Prometheus provides mechanisms of doing service discovery (e.g. `*_sd_config`). The chose of the exact
mechanism depends on the choosen way of scaling.

> 2. What graphs about the service would you plot in a dashboard builder like Grafana?
     Ideally, you can come up with PromQL expressions for them.

TODO

> 3. What would you alert on? What would be the urgency of the various alerts?
     Again, it would be great if you could formulate alerting conditions with PromQL.

TODO

> 4. If you were in control of the microservice, which exported metrics would you add or modify next?

As it was said in the original task, it's a good idea to expose metrics in a format that could be directly
consumed by a metrics aggrerator. This reduces the amount of intermediate componets in the whole system.

In addition to metrics related to service's business logic, one might want to expose data related
to specific programming languages, e.g. number of goroutines of GC timmings for Golang, as well
as system metrics like memory and CPU utilization.

Depening on what exactly the application is doing, it's might be a good idea to provide data related
to the internals of request processing. Timmings for RPC calls and DB querying are the candidates here.

[1]: https://github.com/moby/moby/releases/tag/v17.05.0-ce-rc1
