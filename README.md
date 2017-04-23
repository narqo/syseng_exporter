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

The following metrics are exported from the syseng-challenge service:

- `syseng_http_requests_total{code="NNN"}` - shows how often each HTTP status code has been served
  during the lifetime of the syseng-challenge binary. It's equivalent to `requestCounters.NNN` stat,
  provided by syseng-challenge.
- `syseng_http_request_duration_seconds_count` - how many requests have been served in total during
  the lifetime of the syseng-challenge binary. Equivalent to `duration.count` stat.
- `syseng_http_request_duration_seconds_sum` - how much total time those requests have taken.
  Equivalent to `duration.sum` stat.
- `syseng_up` - indicates whether the last scrap succeeded.

Prometheus is running at http://localhost:9090, open the address in web browser.