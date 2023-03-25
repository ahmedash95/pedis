# Pedis

A sleek, Golang-powered learning adventure inspired by Redis

## What is Pedis?

Pedis is a Redis clone written in Golang. It is a learning project for me to learn Redis protocol. It is not intended to be a production-ready Redis clone.


### Build

TBD

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

	server := pedis.NewServer()
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

## License
MIT
