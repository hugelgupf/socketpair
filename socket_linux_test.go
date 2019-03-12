// +build go1.12

package socketpair

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestHanging(t *testing.T) {
	pc1, _, err := PacketSocketPair()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b := make([]byte, 1)
		log.Printf("reading...")
		pc1.ReadFrom(b)
		log.Printf("reading returned")
	}()

	time.Sleep(2 * time.Second)
	log.Printf("closing...")
	pc1.Close()
	log.Printf("closed")
	wg.Wait()
}
