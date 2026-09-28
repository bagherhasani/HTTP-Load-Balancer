# HTTP Load Balancer

Go program that sits in front of a few HTTP servers. Each request goes to the next backend (round-robin). If a server fails a health check, it stops getting traffic until it is healthy again.

Linux. Docker for running this next to a couple of backends on one machine.

```
cmd/lb              the process you start
internal/backend    one backend
internal/pool       round-robin + who is healthy
demo                fake servers for trying it locally
configs             listen port, backend list, health check
deploy              Docker
```

No code in here yet.
