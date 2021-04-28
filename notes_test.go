package notepet

import (
	"fmt"
	"testing"
	"time"
)

func TestHashGenerator(t *testing.T) {
	stamp := time.Now()
	h1 := generateID(Note{Title: "hello", TimeStamp: stamp})
	h2 := generateID(Note{Title: "hello", TimeStamp: time.Now()})
	h3 := generateID(Note{Title: "hello", TimeStamp: stamp})
	if h1 == h2 {
		fmt.Println("hash is not unique")
		t.Fail()
	}
	if h1 != h3 {
		fmt.Println("hash should be the same for one note")
		t.Fail()
	}
}
