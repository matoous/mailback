package mail

import (
	"errors"
	"net"
	"sort"
	"strings"
)

// MXRecordForHost obtains best available MX records for given host.
func MXRecordForHost(host string) (string, error) {
	recs, err := net.LookupMX(host)
	if err != nil {
		return "", nil
	}
	if len(recs) < 1 {
		return "", errors.New("no MX records found")
	}
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Pref > recs[j].Pref
	})
	// mx records end with '.' so trim it
	rec := strings.TrimRight(recs[0].Host, ".")
	return rec, nil
}
