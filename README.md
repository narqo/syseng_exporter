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
downloads supporting images with syseng-challenge service and Prometheus and runs everything.

Prometheus has been started on http://localhost:9090.

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
They can be calculated using total / summary values, so it's unnessecary to export them.

`rate(syseng_http_requests_total{code="200"}[1m])` calculates the per-second rate (QPS) for requests
with status code 200 over the last 1 minute.

See "[Drop less useful statistics](https://prometheus.io/docs/instrumenting/writing_exporters/#drop-less-useful-statistics)"
section from Prometheus own documentation on writing exporters.

## Bonus Questions

> 1. What are good ways of deploying hundreds of instances of our simulated service?
     How would you deploy your exporter? And how would you configure Prometheus to monitor them all?

[docker-stack](https://docs.docker.com/engine/reference/commandline/stack/) might be helpful for scaling.
The exporter could be run as a sidecar container, so each syseng-challege instance is linked to the exporter
instance in a 1-to-1 fashion.

Different solutions for container orchestration provide the concept of grouping for logically tied
services (syseng-challenge and the exporter in our example). For example [Nomad](https://www.nomadproject.io)
operates with groups, while [Kubernetes](https://kubernetes.io) uses pods.

Prometheus provides mechanisms of doing service discovery (e.g. `*_sd_config` configuration directive).
The chose of the exact mechanism depends on the choosen infrastructure.

> 2. What graphs about the service would you plot in a dashboard builder like Grafana?
     Ideally, you can come up with PromQL expressions for them.

Per status QPS:

~~~
rate(syseng_http_requests_total[5m])
~~~

Average request time:

~~~
rate(syseng_http_request_duration_seconds_sum[5m]) / rate(syseng_http_request_duration_seconds_count[5m])
~~~

Per instance SLA via good, e.g. non 5xx, HTTP status codes:

~~~
sum(rate(syseng_http_requests_total{code!~"^5..$"}[5m])) by (instance)
/
sum(rate(syseng_http_requests_total[5m])) by (instance)
~~~

> 3. What would you alert on? What would be the urgency of the various alerts?
     Again, it would be great if you could formulate alerting conditions with PromQL.

Alert if syseng hasn't exposed it's stats to the exporter withing last minute. The alert is urgent,
as it might mean that a syseng instance is down.

~~~
ALERT InstanceDown
  IF syseng_up == 0
  FOR 1m
  LABELS { severity = "urgent" }
  ANNOTATIONS {
    summary = "{{ $labels.instance }}: down",
    description = "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 1 minutes.",
  }
~~~

Alert for a high rate of failed requests.

~~~
ALERT HighErrorRate
  IF sum(rate(syseng_http_requests_total{code="500"}[5m])) / sum(rate(syseng_http_requests_total[5m])) > 0.01
  FOR 1m
  LABELS { severity = "critical" }
  ANNOTATIONS {
    summary = "{{ $labels.instance }}: high error rate",
    description = "{{ $labels.instance }} of job {{ $labels.job }} has high error rate ({{ $value }}).",
  }
~~~

> 4. If you were in control of the microservice, which exported metrics would you add or modify next?

As it was said in the original task, it's a good idea to expose metrics in a format that could be directly
consumed by a metrics aggregator. This reduces the number of intermediate componets in the whole system.

In addition to metrics related to service's business logic, one might want to expose data related
to specific programming languages, e.g. number of goroutines of GC timmings for Golang, as well
as system metrics like memory and CPU utilization.

Depending on what exactly the application is doing, it might be a good idea to provide data related
to the internals of request processing. Timings for RPC calls and DB querying are the candidates here.

[1]: https://github.com/moby/moby/releases/tag/v17.05.0-ce-rc1
