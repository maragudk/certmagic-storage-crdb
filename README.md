# certmagic-storage-crdb

[![GoDoc](https://godoc.org/github.com/maragudk/certmagic-storage-crdb?status.svg)](https://godoc.org/github.com/maragudk/certmagic-storage-crdb)

An implementation of [certmagic's Storage interface](https://pkg.go.dev/github.com/caddyserver/certmagic#Storage) for CockroachDB.

See [tables.sql](tables.sql) for the expected tables in the database.

## Usage

```go
package main

import (
	"context"
	"log"

	"github.com/caddyserver/certmagic"
	crdb "github.com/maragudk/certmagic-storage-crdb"
)

func main() {
	s := crdb.New(crdb.Options{
		User:     "certmagic",
		Host:     "localhost",
		Port:     26257,
		Database: "certmagic",
		Cert:     "path/to/cert",
		Key:      "path/to/key",
		RootCert: "path/to/root/cert",
	})
	certmagic.Default.Storage = s
	if err := s.Connect(context.Background()); err != nil {
		log.Fatalln("Error connecting to storage:", err)
	}

	err := certmagic.HTTPS([]string{"example.com"}, nil)
}
```

## Limitations

`List` and `Stat` behave a bit differently than the default filesystem implementation,
in that they only support terminal keys, i.e. keys that have data.
The filesystem implementation, in contrast, has directories that can be traversed.
This currently has no implications on certmagic, so you can safely use this implementation.

This may change in a later version, if needed.
