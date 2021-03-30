package notepet

import (
	"fmt"
	"testing"
	"time"
)

func TestHashGenerator(t *testing.T) {
	stamp := time.Now()
	h1 := Note{Title: "hello", TimeStamp: stamp}.GenerateID()
	h2 := Note{Title: "hello", TimeStamp: time.Now()}.GenerateID()
	h3 := Note{Title: "hello", TimeStamp: stamp}.GenerateID()
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
