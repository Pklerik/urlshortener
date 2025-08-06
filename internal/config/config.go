package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StartupFalgs - app startup flags.
type StartupFalgs struct {
	ServerAddress   Address
	AddressShortURL string
	Timeout         Timeout
}

// Address base struct.
type Address struct {
	Protocol string
	Host     string
	Port     int
}

// String provide string representation of Address.
func (a *Address) String() string {
	if a.Protocol == "" {
		a.Protocol = "http"
	}

	if a.Host == "" {
		a.Host = "localhost"
	}

	if a.Port == 0 {
		a.Port = 8080
	}

	return fmt.Sprintf("%s://%s:%d", a.Protocol, a.Host, a.Port)
}

// Set parse Address from string.
func (a *Address) Set(flagValue string) error {
	flagValueMod := flagValue
	if strings.Contains(flagValue, "http") {
		a.Protocol = strings.Split(flagValue, ":")[0]
	}

	if len(strings.Split(flagValue, "://")) == 2 {
		flagValueMod = strings.Split(flagValue, "://")[1]
	}

	if len(strings.Split(flagValueMod, "/")) > 1 {
		flagValueMod = strings.Split(flagValueMod, "/")[0]
	}

	a.Host = strings.Split(flagValueMod, ":")[0]

	port, err := strconv.Atoi(strings.Split(flagValueMod, ":")[1])
	if err != nil {
		return fmt.Errorf("can't set Address Port for %s: %w", flagValue, err)
	}

	if a.Protocol == "" {
		a.Protocol = "http"
	}

	if a.Host == "" {
		a.Host = "localhost"
	}

	a.Port = port
	if a.Port == 0 {
		a.Port = 8080
	}

	return nil
}

// Timeout - contains timeout in float64 of seconds.
type Timeout struct {
	Seconds float64
}

// String - represent timeout in seconds.
func (t *Timeout) String() string {
	return fmt.Sprintf("timeout seconds: %f", t)
}

// Set - set Timeout from string.
func (t *Timeout) Set(flagValue string) error {
	millisecondsR := regexp.MustCompile(`[0-9]*ms`)
	secondR := regexp.MustCompile(`[0-9]*s`)
	minutesR := regexp.MustCompile(`[0-9]*m`)

	milliseconds, _ := strconv.Atoi(millisecondsR.FindString(flagValue))

	seconds, _ := strconv.Atoi(secondR.FindString(flagValue))

	minutes, _ := strconv.Atoi(minutesR.FindString(flagValue))
	if milliseconds+seconds+minutes == 0 {
		minutes = 10
	}

	t.Seconds = float64(milliseconds)/1000 + float64(seconds) + (float64(minutes) * 60)
	if t.Seconds == 0 {
		t.Seconds = 600
	}

	return nil
}
