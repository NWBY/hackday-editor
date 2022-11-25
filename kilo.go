package main

import "os"

func main() {
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		// ignore
	}
	os.Exit(0)
}
