package httplog

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// For unit test only.
var timeNow func() time.Time

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// The FieldLogger interface generalizes structured logging used by httplog.
type FieldLogger interface {
	WithFields(fields Fields) FieldLogger
	Log(args ...interface{})
}

// Config is a configuration for httplog.
type Config struct {
	// RequestFields specifies request fields that should be logged when request is received (before server handling).
	RequestFields []RequestField

	// ResponseFields specifies response fields that should be logged when response is returned/redirected
	// (right after server handling).
	ResponseFields []ResponseField
	// ResponseReqFields specifies request fields that should be logged when response is returned/redirected
	// (right after server handling). It is useful if you want to log only once per request. (common logging technique)
	ResponseReqFields []RequestField
}

// Logger is an instance for httplog to register middleware and wrap response.
type Logger struct {
	// Logger to use internally.
	// TODO(bplotka): Add default FieldLogger (using Bplotka/sgl e.g)
	logger FieldLogger
	cfg    Config
}

func New(logger FieldLogger, cfg Config) *Logger {
	timeNow = time.Now
	return &Logger{
		logger: logger,
		cfg:    cfg,
	}
}

// RegisterMiddleware registers handler that will log request at the beginning and served response at the request end.
func RegisterMiddleware(logger FieldLogger, cfg Config) func(http.Handler) http.Handler {
	l := New(logger, cfg)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log specified RequestFields now.
			l.RequestHandler()(w, r)
			h.ServeHTTP(
				// Log specified ResponseFields and ResponseReqFields on Response Write or Redirect.
				l.WrapResponse(w, r),
				r,
			)
		})
	}
}

func (l *Logger) RequestHandler() func(w http.ResponseWriter, r *http.Request) {
	if len(l.cfg.RequestFields) == 0 {
		return func(_ http.ResponseWriter, _ *http.Request) {}
	}

	return func(_ http.ResponseWriter, r *http.Request) {
		f := Fields{}
		for _, field := range l.cfg.RequestFields {
			v := field.computeValue(timeNow, r)
			if v == "" {
				continue
			}
			f[string(field)] = v
		}

		logger := l.logger
		if len(f) != 0 {
			logger = logger.WithFields(f)
		}
		logger.Log("Received HTTP request")
	}
}

func (l *Logger) WrapResponse(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	return &responseLogger{
		writer:  w,
		req:     r,
		cfg:     l.cfg,
		logger:  l.logger,
		timeNow: timeNow,
	}
}

// RequestField is a log field that can be deducted from http.Request.
type RequestField string

const (
	ReqTimeField  = RequestField("req_time")
	IDField       = RequestField("req_id")
	RemoteIPField = RequestField("req_remote_ip")
	HostField     = RequestField("req_host")
	URIField      = RequestField("req_uri")
	ReqArgsField  = RequestField("req_args")
	MethodField   = RequestField("req_method")
	PathField     = RequestField("req_path")
	BytesInField  = RequestField("req_bytes_in")
	AuthField     = RequestField("req_auth_header")
)

// DefaultRequestFields is a list for recommended configuration of request fields.
var DefaultRequestFields = []RequestField{
	ReqTimeField,
	IDField,
	RemoteIPField,
	HostField,
	ReqArgsField,
	MethodField,
	PathField,
	BytesInField,
	AuthField,
}

// ResponseField is a log field that can be deducted from response.
// It is done by wrapping http.ResponseWriter.
type ResponseField string

const (
	StatusField       = ResponseField("res_status")
	BytesOutField     = ResponseField("res_bytes_out")
	ResTimeField      = ResponseField("res_time")
	ContentTypeField  = ResponseField("res_content_type")
	LocationField     = ResponseField("res_location")
	LocationArgsField = ResponseField("res_location_args")
	LocationHostField = ResponseField("res_location_host")
)

// DefaultResponseFields is a list for recommended configuration of response fields.
var DefaultResponseFields = []ResponseField{
	StatusField,
	BytesOutField,
	ResTimeField,
	ContentTypeField,
	LocationArgsField,
	LocationHostField,
}

// DefaultReqResConfig is configuration for logging one entry when request is received and one when response is written.
func DefaultReqResConfig() Config {
	return Config{
		RequestFields:  DefaultRequestFields,
		ResponseFields: DefaultResponseFields,
	}
}

// DefaultResponseOnlyConfig is configuration for logging only an entry when response is written.
func DefaultResponseOnlyConfig() Config {
	return Config{
		ResponseReqFields: DefaultRequestFields,
		ResponseFields:    DefaultResponseFields,
	}
}

func formatCompactArgs(argQuery string) string {
	argElems := strings.Split(argQuery, "&")
	argsOnly := []string{}
	for _, argElem := range argElems {
		a := strings.Split(argElem, "=")
		if len(a) == 0 || a[0] == "" {
			continue
		}
		argsOnly = append(argsOnly, a[0])
	}
	if len(argsOnly) == 0 {
		return ""
	}
	for i := range argsOnly {
		argsOnly[i] = fmt.Sprintf("%s=...", argsOnly[i])
	}
	return strings.Join(argsOnly, "&")
}

func (f RequestField) computeValue(timeNow func() time.Time, req *http.Request) string {
	switch f {
	case ReqTimeField:
		return timeNow().Format(time.RFC3339)
	case IDField:
		return req.Header.Get("X-Request-ID")
	case RemoteIPField:
		ra := req.RemoteAddr
		if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
			ra = ip
		} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
			ra = ip
		} else {
			ra, _, _ = net.SplitHostPort(ra)
		}
		return ra
	case HostField:
		return req.Host
	case URIField:
		return req.RequestURI
	case ReqArgsField:
		// Parse all form values.
		req.FormValue("")

		return formatCompactArgs(req.Form.Encode())
	case MethodField:
		return req.Method
	case PathField:
		p := req.URL.Path
		if p == "" {
			p = "/"
		}
		return p
	case BytesInField:
		cl := req.Header.Get("Content-Length")
		if cl == "" {
			cl = "0"
		}
		return cl
	case AuthField:
		return req.Header.Get("Authorization")
	default:
		return "not supported"
	}
}

func (f ResponseField) computeValue(timeNow func() time.Time, res *responseLogger) string {
	switch f {
	case StatusField:
		return fmt.Sprintf("%d", res.status)
	case BytesOutField:
		return fmt.Sprintf("%d", res.size)
	case ResTimeField:
		return timeNow().Format(time.RFC3339)
	case ContentTypeField:
		return res.Header().Get("Content-Type")
	case LocationField:
		return res.Header().Get("Location")
	case LocationArgsField:
		splittedQuery := strings.Split(res.Header().Get("Location"), "?")
		if len(splittedQuery) != 2 {
			return ""
		}
		return formatCompactArgs(splittedQuery[1])
	case LocationHostField:
		splittedQuery := strings.Split(res.Header().Get("Location"), "?")
		if len(splittedQuery) < 1 {
			return ""
		}
		return splittedQuery[0]
	default:
		return "not supported"
	}
}

// responseLogger is light wrapper of ResponseWriter and Flusher to support logging on response.
type responseLogger struct {
	writer    http.ResponseWriter
	req       *http.Request
	cfg       Config
	logger    FieldLogger
	status    int
	size      int64
	committed bool
	logged    bool

	timeNow func() time.Time
}

// Header wraps writer Header method.
// See [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter)
func (r *responseLogger) Header() http.Header {
	return r.writer.Header()
}

// WriteHeader wraps writer WriteHeader method.
// See [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter)
func (r *responseLogger) WriteHeader(code int) {
	if r.committed {
		return
	}
	r.status = code
	r.writer.WriteHeader(code)
	r.committed = true

	if r.Header().Get("Location") != "" {
		r.log([]byte{})
	}
}

// Write wraps writer Write method.
// See [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter)
func (r *responseLogger) Write(b []byte) (n int, err error) {
	if !r.committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.writer.Write(b)
	r.size += int64(n)

	r.log(b)
	return
}

// parse Body into structured log entry in best effort manner and only for supported content type.
func (r *responseLogger) parseBody(b []byte) FieldLogger {
	switch r.Header().Get("Content-Type") {
	case "application/json":
		fallthrough
	case "application/json;charset=UTF-8":
		return r.parseJSON(b)
	}
	return r.logger
}

func (r *responseLogger) parseJSON(b []byte) FieldLogger {
	// TODO(bplotka): Add best effort parse.
	return r.logger
}

func (r *responseLogger) log(b []byte) {
	if r.logged {
		return
	}
	r.logged = true
	logger := r.parseBody(b)

	f := Fields{}
	for _, field := range r.cfg.ResponseReqFields {
		v := field.computeValue(r.timeNow, r.req)
		if v == "" {
			continue
		}
		f[string(field)] = v
	}

	for _, field := range r.cfg.ResponseFields {
		v := field.computeValue(r.timeNow, r)
		if v == "" {
			continue
		}
		f[string(field)] = v
	}

	if len(f) != 0 {
		logger = logger.WithFields(f)
	}

	if r.Header().Get("Location") != "" {
		logger.Log("Redirecting HTTP request")
	} else {
		logger.Log("Responding to HTTP request")
	}
}
