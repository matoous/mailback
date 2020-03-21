package mail

import (
	"strings"
)

// Host returns the email host.
func Host(addr string) string {
	li := strings.LastIndex(addr, "@")
	if li < 0 || li > len(addr)-1 {
		return ""
	}
	return addr[li+1:]
}

// User returns the email user.
func User(addr string) string {
	li := strings.LastIndex(addr, "@")
	if li < 0 {
		return ""
	}
	return addr[:li]
}
