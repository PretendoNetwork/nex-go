package main

import "sync"

var wg sync.WaitGroup

func main() {
	wg.Add(3)

	go startAuthenticationServer()
	go startSecureServer()
	go startHPPServer()

	wg.Wait()
}
