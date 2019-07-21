# go-auto-backends
auto registers service ip:port+weight to backends list with health checks. ( kind of autonomous system )


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
    b := autobackends.New(&PubSub{conn}, "root-area", me, weight)
    b.Start("root-area", me, weight, 1)
    b.Start("sub-area", me, weight*10, 1)
    b.Start("sub-area1", me, weight*100, 1)
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
    b := autobackends.New(&PubSub{conn},"root-area", "my-hostname:port", 1000)

    // wait 1s while backends registers
    time.Sleep( 2 * time.Second )

    for {
        backend, _ := b.Get()
        println("backend", backend)
        time.Sleep( 1 * time.Second)
    }

    runtime.Goexit()
    
}
```