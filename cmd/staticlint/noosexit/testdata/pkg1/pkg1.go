package pkg1

import (
	"fmt"
	"os"
)

func a(i int) (int, error) {
	return i * 2, nil
}

func errCheckFunc() {
	fmt.Println(a(5))
	os.Exit(1)
}
