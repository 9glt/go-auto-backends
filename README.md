# go-auto-backends
auto registers service ip:port+weight to backends list with health checks


server1 
```go
package main 

import (
    autobackends "github.com/9glt/go-auto-backends"
)

type  PuBsub struct {
    /// implmenet autobackends.PubSub interface
}

func main() {
    weight := 10
    me := "192.168.100.100:80"
    b := backends.New(&PubSub{conn}, me, weight)
    b.Start(me, 1)
    runtime.Goexit()
}
```


client1
```go
package main 

import (
    autobackends "github.com/9glt/go-auto-backends"
)

type  PuBsub struct {
    /// implmenet autobackends.PubSub interface
}

func main() {
    b := backends.New(&PubSub{conn}, *flagMe, *flagW)

    // wait 1s while backends registers
    time.Sleep( 2 * time.Second )

    for {
        backend := b.Get()
        println("backend", backend)
        time.Sleep( 1 * time.Second)
    }

    runtime.Goexit()
    
}
```