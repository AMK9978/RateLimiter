# Rate Limiter
Rate limiter in Go

## Get started

Docker:
```
git clone git@github.com:AMK9978/RateLimiter.git
docker-compose up --build -d
```
or simply run a Redis service locally or as a container on a port (default is
6379). Afterward, run the app:
```
go run cmd/ratelimiter/main.go
```
Then send a request:
```
curl http://localhost:8080/limit?userID=1&window_duration=5&limit=3
```
You receive `Request allowed` initially with `200` status code. If you
repeat this within the window duration, you will get `Rate limit exceeded`
with `429` status code`.

## Architecture
The project's structure follows the common pattern of Golang projects. Apart
from the `cmd`, the algorithms and connection to Redis reside in the 
`internal` package. The project follows SOLID principles to enable users
to easily extend the functionalities such as connecting to other DBs, adding
new algorithms, etc. Additionally, the app enjoys circuit breaker pattern
to avoid error cascading. Plus, to control 

In the `internal` directory, there are `config` for essential configs of
the app like port, `limiter` which contains the logic of handling, checking,
and performing sliding window and leaky bucket algorithms, `logger` which
includes the central settings for the app's logger, and `routes` providing
two routes of the application for leaky bucket and sliding window, which can
be called by external services. Lastly, there are both functional tests and 
unit tests for both sliding window and leaky bucket. 

Apart from that, there is a `limiter_test` containing a banchmark test for 
the application.

This stateless system can easily be replicated and it employs `distributed 
locking` based on the `userID` to handle concurrency. The
app includes a RedisInterface to be extended by different detailed implementations.
The app contains `RedisClient` and `MockRedisClient`, but a Redis cluster
connector can easily be added to the program as well.


## Further steps
Although this app is scalable, but Redis can become a hotspot under extremely
heavy loads. Therefore, adding a cluster connector, which is very similar to
the current RedisClient except having a list of nodes and balancing between
them by, for example, a hash-based algorithm, is needed. 


## Notes
The users can run replicates of this application without any problem to 
simulate a distributed system by either running in a kubernetes cluster or
using `docker service scale`. Some comments are added to the code to help
better understanding.