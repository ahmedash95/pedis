# Pedis

A sleek, Golang-powered learning adventure inspired by Redis

## What is Pedis?

Pedis is a Redis clone written in Golang. It is a learning project for me to learn Redis protocol. It is not intended to be a production-ready Redis clone.


### Build

Pedis does not have any external dependencies. To build it, simply run:

```bash
$ go build -o pedis-server && ./pedis-server
```

## Usage

- Initialize a Pedis server

```go
package main

import (
	"ahmedash95/pedis/pedis"
	"fmt"
)

func main() {
	fmt.Println("Listening on port 6379")

	config := &pedis.Config{
		EnableAof: true,
	} // to enable AOF persistence or set config to nil

	server := pedis.NewServer(config)
	server.ListenAndServe("0.0.0.0:6379")
}
```

- Connect to the server using a Redis client

```bash
$ redis-cli -p 6379
```

## Supported Commands

- [x] GET
- [x] SET
- [x] DEL
- [x] EXISTS
- [ ] EXPIRE
- [ ] TTL
- [x] HSET
- [x] HGET
- [x] HGETALL
- [x] HDEL
- [x] HLEN
- [x] HKEYS
- [x] HVALS

## Persistence
At the moment, Pedis supports AOF persistence. It is disable by default. To enabled it, set `EnableAof` to `true`
in the config. the policy for now is to append to the AOF file every second.

## License
MIT
