package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Fprintln(os.Stderr, "i am just test plugin. i don't function")
	time.Sleep(10 * time.Millisecond) // Hack to workaround the stderr message not being captured by the time the malformed stdout is processed.
	fmt.Fprintln(os.Stdout, "i am a little teapot")
}
