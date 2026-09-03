package pkgmgmt

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
)

// IsRetryableError reports whether err is a transient network error worth
// retrying, e.g. a connection reset, refused connection, or timeout.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return true
	}

	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	msg := err.Error()
	for _, s := range []string{"TLS handshake timeout", "connection reset by peer", "connection refused"} {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}

// IsRetryableStatus reports whether an HTTP status code represents a
// transient failure worth retrying.
func IsRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}
