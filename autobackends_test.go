package autobackends

import (
	"testing"
	"time"

	nserver "github.com/nats-io/nats-server/server"
)

func TestMain(t *testing.T) {
	srv := nserver.New(&nserver.Options{
		Host: "127.0.0.1",
		Port: 4334,
	})
	go func() {
		srv.Start()
	}()
	time.Sleep(time.Second)

	routes := NewTable()

	routes.Add("127.0.0.1", "nop", 10)
	routes.Add("127.0.0.2", "127.0.0.1", 10)
	routes.Add("127.0.0.3", "127.0.0.2", 10)
	routes.Add("127.0.0.4", "127.0.0.3", 10)
	routes.Add("127.0.0.5", "127.0.0.4", 10)

	b1 := New(Nats("nats://127.0.0.1:4334", "autobackends.live"), routes.List("127.0.0.1", nil), "127.0.0.1")
	b2 := New(Nats("nats://127.0.0.1:4334", "autobackends.live"), routes.List("127.0.0.2", nil), "127.0.0.2")
	b3 := New(Nats("nats://127.0.0.1:4334", "autobackends.live"), routes.List("127.0.0.3", nil), "127.0.0.3")
	b4 := New(Nats("nats://127.0.0.1:4334", "autobackends.live"), routes.List("127.0.0.4", nil), "127.0.0.4")
	b5 := New(Nats("nats://127.0.0.1:4334", "autobackends.live"), routes.List("127.0.0.5", nil), "127.0.0.5")

	time.Sleep(time.Second)
	_, err := b1.GetCached()
	if err == nil {
		t.Fatal()
	}
	be2, err := b2.GetCached()
	if err != nil {
		t.Fatal()
	}
	be3, err := b3.GetCached()
	if err != nil {
		t.Fatal()
	}
	be4, err := b4.GetCached()
	if err != nil {
		t.Fatal()
	}

	if be2 != "127.0.0.1" {
		t.Fatal(be2)
	}
	if be3 != "127.0.0.2" {
		t.Fatal(be3)
	}
	if be4 != "127.0.0.3" {
		t.Fatal(be4)
	}
	b2.Stop()
	time.Sleep(3 * time.Second)

	be3, err = b3.GetCached()
	if err != nil {
		t.Fatal()
	}
	be4, err = b4.Get()
	if err != nil {
		t.Fatal()
	}

	if be3 != "127.0.0.1" {
		t.Fatal(be3)
	}
	if be4 != "127.0.0.3" {
		t.Fatal(be4)
	}
	b3.Stop()
	time.Sleep(3 * time.Second)

	be4, err = b4.GetCached()
	if err != nil {
		t.Fatal()
	}

	if be4 != "127.0.0.1" {
		t.Fatal(be4)
	}
	b1.Stop()
	b4.Stop()
	time.Sleep(3 * time.Second)
	be5, err := b5.Get()
	if err == nil {
		t.Fatal(be5)
	}
	srv.Shutdown()
}
