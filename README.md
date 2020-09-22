# certmagic-storage-crdb

An implementation of [certmagic's Storage interface](https://pkg.go.dev/github.com/caddyserver/certmagic#Storage) for CockroachDB.

See [tables.sql](tables.sql) for the expected tables in the database.

## Limitations

`List` and `Stat` behave a bit differently than the default filesystem implementation,
in that they only support terminal keys, i.e. keys that have data.
The filesystem implementation, in contrast, has directories that can be traversed.

This may change in a later version, if needed.

## Usage

```go
package main

import (
	crdb "github.com/maragudk/certmagic-storage-crdb"
)

func main() {
    storage := crdb.New(crdb.Options{
        User:     "certmagic",
        Host:     "localhost",
        Port:     26257,
        Database: "certmagic",
        Cert:     "path/to/cert",
        Key:      "path/to/key",
        RootCert: "path/to/root/cert",
    })
    // use storage in certmagic or caddy
}
```
