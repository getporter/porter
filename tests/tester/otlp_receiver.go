package tester

import (
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	// OTLPProtocolGRPC is the telemetry protocol for sending traces over grpc.
	OTLPProtocolGRPC = "grpc"

	// OTLPProtocolHTTP is the telemetry protocol for sending traces over http.
	OTLPProtocolHTTP = "http/protobuf"
)

// TestOTLPReceiver is a temporary OpenTelemetry endpoint that collects the
// trace data sent to it, and is stopped when the test completes.
type TestOTLPReceiver struct {
	coltracepb.UnimplementedTraceServiceServer

	t *testing.T

	// protocol that the receiver accepts, either grpc or http/protobuf
	protocol string

	// endpoint is the host:port that the receiver is listening on
	endpoint string

	// stop shuts down the underlying server
	stop func()

	mu            sync.Mutex
	resourceSpans []*tracepb.ResourceSpans
}

// ReceivedSpan is a span collected by a TestOTLPReceiver.
type ReceivedSpan struct {
	*tracepb.Span

	// ServiceName is the name of the service that exported the span.
	ServiceName string
}

// TraceID returns the hex encoded trace id of the span.
func (s ReceivedSpan) TraceID() string {
	return hex.EncodeToString(s.GetTraceId())
}

// StartTestOTLPReceiver runs an OpenTelemetry endpoint for the specified
// protocol (grpc or http/protobuf) inside the test process, and configures
// every porter command subsequently run by the Tester to send its traces to it.
// The receiver is cleaned up by default when the test completes.
//
// The receiver listens on localhost, so it does not receive the traces
// exported from inside a bundle. Use StartBundleTestOTLPReceiver for that.
func (t Tester) StartTestOTLPReceiver(protocol string) *TestOTLPReceiver {
	return t.startTestOTLPReceiver(protocol, "127.0.0.1")
}

// StartBundleTestOTLPReceiver is like StartTestOTLPReceiver, except that the
// receiver listens on the gateway of the default docker network, so that it
// also receives the traces exported from inside a bundle.
// The test is skipped when the host can't listen on that address, e.g. when
// docker runs in a virtual machine.
func (t Tester) StartBundleTestOTLPReceiver(protocol string) *TestOTLPReceiver {
	gateway, err := shx.OutputE("docker", "network", "inspect", "bridge", "--format", "{{(index .IPAM.Config 0).Gateway}}")
	require.NoError(t.T, err, "Could not determine the gateway of the default docker network")

	lis, err := net.Listen("tcp", net.JoinHostPort(gateway, "0"))
	if err != nil {
		t.T.Skipf("Skipping because the host can't listen on the gateway of the default docker network, %s: %s", gateway, err)
	}
	require.NoError(t.T, lis.Close())

	return t.startTestOTLPReceiver(protocol, gateway)
}

func (t Tester) startTestOTLPReceiver(protocol string, host string) *TestOTLPReceiver {
	r := &TestOTLPReceiver{t: t.T, protocol: protocol}

	lis, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	require.NoError(t.T, err, "Could not listen on a port for the temporary otlp receiver")
	r.endpoint = lis.Addr().String()

	switch protocol {
	case OTLPProtocolGRPC:
		srv := grpc.NewServer()
		coltracepb.RegisterTraceServiceServer(srv, r)
		go func() {
			_ = srv.Serve(lis)
		}()

		r.stop = srv.Stop
	case OTLPProtocolHTTP:
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/traces", r.handleHTTPExport)
		srv := httptest.NewUnstartedServer(mux)
		_ = srv.Listener.Close()
		srv.Listener = lis
		srv.Start()

		r.stop = srv.Close
	default:
		_ = lis.Close()
		require.Failf(t.T, "invalid otlp protocol", "unsupported otlp protocol %q", protocol)
	}

	// Automatically stop the receiver when the test is done
	t.T.Cleanup(r.Close)

	t.SetEnv("PORTER_TELEMETRY_ENABLED", "true")
	t.SetEnv("PORTER_TELEMETRY_PROTOCOL", protocol)
	t.SetEnv("PORTER_TELEMETRY_ENDPOINT", r.endpoint)
	t.SetEnv("PORTER_TELEMETRY_INSECURE", "true")

	return r
}

// String prints the receiver's endpoint.
func (r *TestOTLPReceiver) String() string {
	return r.endpoint
}

// Endpoint is the host:port that the receiver is listening on.
func (r *TestOTLPReceiver) Endpoint() string {
	return r.endpoint
}

// Close stops the receiver.
func (r *TestOTLPReceiver) Close() {
	if r.stop != nil {
		r.stop()
		r.stop = nil
	}
}

// Export implements the grpc TraceService, collecting the exported spans.
func (r *TestOTLPReceiver) Export(_ context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	r.collect(req)
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

// handleHTTPExport implements the http/protobuf TraceService, collecting the exported spans.
func (r *TestOTLPReceiver) handleHTTPExport(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var body io.Reader = req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gz.Close()
		body = gz
	}

	data, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exportReq coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(data, &exportReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.collect(&exportReq)

	resp, err := proto.Marshal(&coltracepb.ExportTraceServiceResponse{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	_, _ = w.Write(resp)
}

func (r *TestOTLPReceiver) collect(req *coltracepb.ExportTraceServiceRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resourceSpans = append(r.resourceSpans, req.GetResourceSpans()...)
}

// AllSpans returns every span that has been received so far.
func (r *TestOTLPReceiver) AllSpans() []ReceivedSpan {
	r.mu.Lock()
	defer r.mu.Unlock()

	var spans []ReceivedSpan
	for _, rs := range r.resourceSpans {
		serviceName := ""
		for _, attr := range rs.GetResource().GetAttributes() {
			if attr.GetKey() == "service.name" {
				serviceName = attr.GetValue().GetStringValue()
			}
		}

		for _, ss := range rs.GetScopeSpans() {
			for _, span := range ss.GetSpans() {
				spans = append(spans, ReceivedSpan{Span: span, ServiceName: serviceName})
			}
		}
	}
	return spans
}

// Spans returns the spans that have been received so far from the specified service.
func (r *TestOTLPReceiver) Spans(serviceName string) []ReceivedSpan {
	var spans []ReceivedSpan
	for _, span := range r.AllSpans() {
		if span.ServiceName == serviceName {
			spans = append(spans, span)
		}
	}
	return spans
}

// RequireSpans fails the test when no spans were received from the specified
// service, otherwise it returns the service's spans.
func (r *TestOTLPReceiver) RequireSpans(serviceName string) []ReceivedSpan {
	r.t.Helper()

	spans := r.Spans(serviceName)
	require.NotEmptyf(r.t, spans, "expected the %s otlp receiver at %s to have received spans from the %s service", r.protocol, r.endpoint, serviceName)
	return spans
}

// RequireNoSpanContains fails the test when the specified value is found
// anywhere in the received trace data: span names, status messages,
// attributes, events or event attributes.
func (r *TestOTLPReceiver) RequireNoSpanContains(value string) {
	r.t.Helper()

	var found []string
	check := func(span ReceivedSpan, location string, text string) {
		if strings.Contains(text, value) {
			found = append(found, fmt.Sprintf("%s span %q: %s", span.ServiceName, span.GetName(), location))
		}
	}
	checkAttributes := func(span ReceivedSpan, location string, attrs []*commonpb.KeyValue) {
		for _, attr := range attrs {
			check(span, fmt.Sprintf("%s attribute %s", location, attr.GetKey()), attr.GetKey())
			// Use the text representation so that we search all value types, including nested lists and maps
			check(span, fmt.Sprintf("%s attribute %s", location, attr.GetKey()), attr.GetValue().String())
		}
	}

	for _, span := range r.AllSpans() {
		check(span, "name", span.GetName())
		check(span, "status message", span.GetStatus().GetMessage())
		checkAttributes(span, "span", span.GetAttributes())
		for _, event := range span.GetEvents() {
			check(span, "event name", event.GetName())
			checkAttributes(span, fmt.Sprintf("event %q", event.GetName()), event.GetAttributes())
		}
	}

	require.Emptyf(r.t, found, "expected %q to not be in the trace data, but it was found in:\n%s", value, strings.Join(found, "\n"))
}
