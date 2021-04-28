package notepet

import (
	"fmt"
	"testing"
)

func TestOpenSQLiteStorage(t *testing.T) {
	if _, err := OpenSQLiteStorage("noexistent"); err == nil {
		fmt.Println("OpenSQLiteStorage creates file when not needed")
		t.Fail()
	}
}
