# go-auto-backends
auto registers service ip:port+weight to backends list with health checks. ( kind of autonomous system )


server1 
```go
package main 

import (
    autobackends "github.com/9glt/go-auto-backends"
    "fmt"
    "flag"
)

type  PuBsub struct {
    /// implmenet autobackends.PubSub interface
}

var (
    flagMe = flag.String("me", "127.0.0.1", "my ip:port")
)

func main() {
    flag.Parse()

    routes := autobackends.NewTable()

    // area begin
    routes.Add("127.0.0.2", "127.0.0.1", 10)
	routes.Add("127.0.0.20", "127.0.0.2", -1) // failover for 5 and 6 and 3 and 4
	// area end

	// area begin
	routes.Add("127.0.0.3", "127.0.0.20", 10)
	routes.Add("127.0.0.4", "127.0.0.3", -1) // failover for 5 and 6
	// area end

	// area begin
	routes.Add("127.0.0.5", "127.0.0.4", 10)
    routes.Add("127.0.0.6", "127.0.0.5", -1)
    // area end

    b := autobackends.New(&PubSub{conn}, routes.List(*flagMe, nil), *flagMe)
    for {
        be, err := b.Get()
        if err != nil {
            println(err.Error())
        } else {
            fmt.Printf("%+v\n", be)
        }
        time.Sleep(time.Second)
    }
    runtime.Goexit()
}
```

```bash
go run main.go -me 127.0.0.1
go run main.go -me 127.0.0.2
go run main.go -me 127.0.0.3
go run main.go -me 127.0.0.4
go run main.go -me 127.0.0.5
go run main.go -me 127.0.0.6
```
turn random node off