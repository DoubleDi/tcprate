# Throttling bandwidth with Go

## Test

```
go test -v -timeout 300s
```


## Example server usage

```
    l, err := net.Listen("tcp", addr)
    if err != nil {
        panic(err)
    }
    l.WithBandwith(10)
    for {
        conn, err := l.Accept()
        if err != nil {
            panic(err)
        }
        wrappedConn := limiter.WrapConn(conn)
        wrappedConn.WithBandwith(5)
        go handleConn(t, wrappedConn)
    }
```