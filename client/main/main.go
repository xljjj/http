package main

import (
	"HTTP/client"
	"fmt"
)

func main() {
	// There can be more than one client
	fmt.Println("HTTP client starts up")
	client.Run("127.0.0.1", 8888)
	fmt.Println("HTTP client closes up")
}
