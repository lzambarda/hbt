package main

import (
	"fmt"
	"os"

	"github.com/lzambarda/hbt/cmd"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
