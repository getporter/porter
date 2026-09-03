package pkgmgmt

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRetryableError(t *testing.T) {
	testcases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "EOF", err: io.EOF, want: true},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, want: true},
		{name: "TLS handshake timeout", err: errors.New("net/http: TLS handshake timeout"), want: true},
		{
			name: "connection reset by peer",
			err:  fmt.Errorf("read tcp 10.1.0.1:1234->1.2.3.4:443: %w", os.NewSyscallError("read", syscall.ECONNRESET)),
			want: true,
		},
		{
			name: "connection refused",
			err:  fmt.Errorf("dial tcp 1.2.3.4:443: %w", os.NewSyscallError("connect", syscall.ECONNREFUSED)),
			want: true,
		},
		{
			name: "net.Error timeout",
			err:  &net.DNSError{IsTimeout: true},
			want: true,
		},
		{name: "not retryable", err: errors.New("boom"), want: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsRetryableError(tc.err))
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	testcases := []struct {
		code int
		want bool
	}{
		{code: http.StatusOK, want: false},
		{code: http.StatusNotFound, want: false},
		{code: http.StatusBadRequest, want: false},
		{code: http.StatusTooManyRequests, want: true},
		{code: http.StatusInternalServerError, want: true},
		{code: http.StatusBadGateway, want: true},
		{code: http.StatusServiceUnavailable, want: true},
		{code: http.StatusGatewayTimeout, want: true},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d", tc.code), func(t *testing.T) {
			assert.Equal(t, tc.want, IsRetryableStatus(tc.code))
		})
	}
}
