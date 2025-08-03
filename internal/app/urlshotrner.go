// Package app contain app function for startup.
package app

import "github.com/Pklerik/urlshortener/internal/router"

// Start - main app function.
func Start() {
	router.StartServer()
}
