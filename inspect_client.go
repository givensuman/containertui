// inspect_client.go is a development utility for inspecting the Docker client API.
// This file is used during development to explore available methods on the Docker client.
// It can be removed in production builds or kept for debugging purposes.
//
// To use: go run inspect_client.go
package main

import (
	"fmt"
	"reflect"

	"github.com/docker/docker/client"
)

func main() {
	c, _ := client.NewClientWithOpts(client.FromEnv)
	t := reflect.TypeOf(c)
	fmt.Printf("Client Type: %s\n", t)
	for i := range t.NumMethod() {
		m := t.Method(i)
		fmt.Printf("Method: %s\n", m.Name)
	}
	fmt.Println("Scan complete")
}
