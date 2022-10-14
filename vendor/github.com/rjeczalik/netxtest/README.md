# netxtest
Package simplifying creation of netx integration tests.

Since this repository is a private one you need to ensure your git configuration allows for `go get`ting it.

This can be done e.g. by rewriting the remote path with the following configuration:

```
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### Usage for `netxtest.LimitListenerTest`

Create a test command for your package.

The test interfaces with your package with `netxtest.LimitListenerFunc`, example implementation:

```go
func myLimitedListener(l net.Listener, limitGlobal, limitPerConn int) net.Listener {
	limited := mypkg.NewListener(l)
	limited.SetLimits(limitGlobal, limitPerConn)
	return limited
}
```

Create a `mypkg.test/main.go` source file with the following content (adjust myLimitedListener accordingly to your packages API):

```go
package main

import (
	"flag"
	"log"
	"net"

	"github.com/me/mypkg"

	"github.com/rjeczalik/netxtest"
)

func myLimitedListener(l net.Listener, limitGlobal, limitPerConn int) net.Listener {
	limited := mypkg.NewListener(l)
	limited.SetLimits(limitGlobal, limitPerConn)
	return limited
}

func main() {
	var test netxtest.LimitListenerTest

	test.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if err := test.Run(myLimitedListener); err != nil {
		log.Fatal(err)
	}
}
```

Build the command:

```
mypkg.test $ go build
```

Run it:

```
mypkg.test $ ./mypg.test -count 5
```

On successful run it will output the following logs:

```
mypkg.test $ ./mypg.test -count 5
2019/09/03 18:53:37 clients: 5
2019/09/03 18:53:37 global limit: 25600 [kB/s], per connection: 256 [kB/s]
2019/09/03 18:53:37 transfer duration: 30s
2019/09/03 18:53:37 expected bandwidth within range (7471104, 8257536) [B] (epsilon=0.05)
2019/09/03 18:53:37 running test ...
2019/09/03 18:54:07 client[0]: transferred bytes: 8126464 [B]
2019/09/03 18:54:07 client[1]: transferred bytes: 8126464 [B]
2019/09/03 18:54:07 client[2]: transferred bytes: 8126464 [B]
2019/09/03 18:54:07 client[3]: transferred bytes: 8126464 [B]
2019/09/03 18:54:07 client[4]: transferred bytes: 8126464 [B]
2019/09/03 18:54:07 OK
```

During testing ensure the following cases pass:

```
$ ./mypg.test -count 100
```
```
$ ./mypg.test -count 100 -limit 0
```
```
$ ./mypg.test -count 100 -limit-conn 0
```
