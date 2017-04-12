package httplog

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// The FieldLogger interface generalizes structured logging used by httplog.
type FieldLogger interface {
	WithFields(fields Fields) FieldLogger
	WithError(err error) FieldLogger

	Log(args ...interface{})
}

type HTTPHandler func(w http.ResponseWriter, r *http.Request)

type Config struct {
	RequestFields []RequestField

	ResponseFields    []ResponseField
	ResponseReqFields []RequestField
}

func DefaultReqResConfig() Config {
	return Config{
		RequestFields: []RequestField{
			ReqTimeField,
			IDField,
			RemoteIPField,
			HostField,
			ReqArgsField,
			MethodField,
			PathField,
			BytesInField,
			AuthField,
		},
		ResponseFields: []ResponseField{
			StatusField,
			BytesOutField,
			ContentTypeField,
		},
	}
}

func DefaultResponseOnlyConfig() Config {
	return Config{
		ResponseReqFields: []RequestField{
			ReqTimeField,
			IDField,
			RemoteIPField,
			HostField,
			URIField,
			MethodField,
			PathField,
			BytesInField,
			AuthField,
		},
		ResponseFields: []ResponseField{
			StatusField,
			BytesOutField,
			ContentTypeField,
			ResTimeField,
		},
	}
}

type Logger struct {
	// Logger to use internally.
	// TODO(bplotka): Add default FieldLogger (using Bplotka/sgl e.g)
	logger FieldLogger
	cfg    Config

	timeNow func() time.Time
}

func New(logger FieldLogger, cfg Config) *Logger {
	return &Logger{
		logger:  logger,
		cfg:     cfg,
		timeNow: time.Now,
	}
}

type RequestField string
type ResponseField string

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

	StatusField      = ResponseField("res_status")
	BytesOutField    = ResponseField("res_bytes_out")
	ResTimeField     = ResponseField("res_time")
	ContentTypeField = ResponseField("res_content_type")
)

func (f RequestField) ComputeValue(timeNow func() time.Time, req *http.Request) string {
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
		req.FormValue("")
		argElems := strings.Split(req.Form.Encode(), "&")
		argsOnly := []string{}
		for _, argElem := range argElems {
			a := strings.Split(argElem, "=")
			if len(a) == 0 {
				continue
			}
			argsOnly = append(argsOnly, a[0])
		}
		if len(argsOnly) == 0 {
			return ""
		}
		for i, _ := range argsOnly {
			argsOnly[i] = fmt.Sprintf("%s=(...)", argsOnly[i])
		}
		return strings.Join(argsOnly, "&")
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

func (f ResponseField) ComputeValue(timeNow func() time.Time, res *http.Response) string {
	switch f {
	case StatusField:
		return res.Status
	case BytesOutField:
		return strconv.FormatInt(res.ContentLength, 10)
	case ResTimeField:
		return timeNow().Format(time.RFC3339)
	case ContentTypeField:
		return res.Header.Get("Content-Type")
	default:
		return "not supported"
	}
}

func (l *Logger) RequestLogger() HTTPHandler {
	if len(l.cfg.RequestFields) == 0 {
		return func(_ http.ResponseWriter, _ *http.Request) {}
	}

	return func(_ http.ResponseWriter, r *http.Request) {
		f := Fields{}
		for _, field := range l.cfg.RequestFields {
			v := field.ComputeValue(l.timeNow, r)
			if v == "" {
				continue
			}
			f[string(field)] = v
		}
		if len(f) == 0 {
			return
		}
		l.logger.WithFields(f).Log("Received HTTP request")
	}
}

func (l *Logger) ResponseLogger() HTTPHandler {
	return func(_ http.ResponseWriter, r *http.Request) {
		f := Fields{}
		for _, field := range l.cfg.ResponseReqFields {
			v := field.ComputeValue(l.timeNow, r)
			if v == "" {
				continue
			}
			f[string(field)] = v
		}

		for _, field := range l.cfg.ResponseFields {
			v := field.ComputeValue(l.timeNow, r.Response)
			if v == "" {
				continue
			}
			f[string(field)] = v
		}

		if len(f) == 0 {
			return
		}

		l.logger.WithFields(f).Log("Responding to HTTP request")
	}
}
