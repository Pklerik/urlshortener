package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type StartupFalgs struct {
	ServerAddress   Address
	AddressShortURL Address
	Timeout         Timeout
}

// BaseURL base address
type Address struct {
	Protocol string
	Host     string
	Port     int
}

func (a *Address) String() string {
	return fmt.Sprintf("%s://%s:%d", a.Protocol, a.Host, a.Port)
}
func (a *Address) Set(flagValue string) error {
	a.Protocol = "http"
	a.Host = "localhost"
	a.Port = 8080
	flagValueMod := flagValue
	if strings.Contains(flagValue, "http") {
		a.Protocol = strings.Split(flagValue, "/")[0]
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
	a.Port = port
	return nil
}

type Timeout struct {
	Seconds int
}

func (t *Timeout) String() string {
	return fmt.Sprintf("timeout seconds: %d", t)
}

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
	t.Seconds = milliseconds/1000 + seconds + (minutes * 60)
	return nil
}
