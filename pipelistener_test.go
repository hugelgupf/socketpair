package socketpair

import (
	"strings"
	"sync"
	"testing"
)

func TestListen(t *testing.T) {
	l := Listen()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			s, err := l.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					t.Fatal(err)
				}
				return
			}

			s.Close()
		}
	}()
	defer wg.Wait()
	defer l.Close()

	client, err := l.Dial()
	if err != nil {
		t.Fatal(err)
	}
	client.Close()
}
