package jiffy

import (
	"fmt"
	"os"
)

var (
	Random *os.File
)

// Generates a random UUID.
func UUID() string {
	b := make([]byte, 16)
	Random.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func init() {
	var err error
	Random, err = os.Open("/dev/urandom")
	if err != nil {
		panic(err)
	}
}
