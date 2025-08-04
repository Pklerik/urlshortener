// Package router provides functionality for setups and startup of server.
package router

import "testing"

func TestStartServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StartServer()
		})
	}
}
