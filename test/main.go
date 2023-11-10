package main

import "sync"

var wg sync.WaitGroup

func main() {
	wg.Add(2)

	go startAuthenticationServer()
	go startSecureServer()

	wg.Wait()
}
