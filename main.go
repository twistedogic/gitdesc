package main

import (
	"flag"
	"os"
)

func main() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cmd := NewCmd(f)
	cmd.Execute(os.Args[1:])
}
