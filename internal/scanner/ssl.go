package scanner

import (
	"crypto/tls"
	"net"
	"time"
)

func checkSSL(domain string) bool {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 3 * time.Second},
		"tcp", domain+":443",
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
