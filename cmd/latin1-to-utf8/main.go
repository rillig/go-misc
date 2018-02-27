package main

import (
	"github.com/rillig/misc/toutf8"
	"io"
	"log"
	"os"
)

func main() {
	_, err := io.Copy(toutf8.NewUtf8Writer(os.Stdout), os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
}
