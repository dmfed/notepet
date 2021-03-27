package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/dmfed/notepet"
)

func TestHashGenerator(t *testing.T) {
	stamp := time.Now()
	h1 := generateID(notepet.Note{Title: "hello", TimeStamp: stamp})
	h2 := generateID(notepet.Note{Title: "hello", TimeStamp: time.Now()})
	h3 := generateID(notepet.Note{Title: "hello", TimeStamp: stamp})
	// fmt.Printf("%v\n%v\n%v\n", h1, h3, h2)
	if h1 == h2 {
		fmt.Println("hash is not unique")
		t.Fail()
	}
	if h1 != h3 {
		fmt.Println("hash should be the same for one note")
		t.Fail()
	}
}
