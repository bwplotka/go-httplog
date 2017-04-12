package httplog

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//go:generate mockery -name FieldLogger -case underscore -inpkg

func testRequest() (*http.Request, error) {
	return http.NewRequest(
		"GET",
		`/some_endpoint?arg1="arg1value"&arg2="arg2value"`,
		nil,
	)
}

func TestLogger_HTTPHandler_EmptyConfig(t *testing.T) {
	mLogger := new(MockFieldLogger)
	l := New(mLogger, Config{})

	req, err := testRequest()
	require.NoError(t, err)
	l.RequestLogger()(nil, req)
	l.ResponseLogger()(nil, req)
}

func TestLogger_HTTPHandler_RequestFieldsLogged(t *testing.T) {
	mLogger := new(MockFieldLogger)
	now := time.Now()
	mLogger.On("WithFields", Fields{
		"req_bytes_in": "0",
		//"req_id":          "",
		//"req_host":        "",
		"req_args": "arg1=(...)|arg2=(...)",
		"req_path": "/some_endpoint",
		"req_time": now.Format(time.RFC3339),
		//"req_remote_ip":   "",
		"req_method": "GET",
		//"req_auth_header": "",
	}).Return(mLogger)
	mLogger.On("Log", []interface{}{"Received HTTP request"})

	l := New(mLogger, DefaultReqResConfig())
	l.timeNow = func() time.Time {
		return now
	}

	req, err := testRequest()
	require.NoError(t, err)
	l.RequestLogger()(nil, req)
}

func TestLogger_HTTPHandler_ResponseFieldsLogged(t *testing.T) {
	mLogger := new(MockFieldLogger)

	mLogger.On("Log", []interface{}{"Received HTTP request"})

	l := New(mLogger, DefaultReqResConfig())

	req, err := testRequest()
	require.NoError(t, err)
	l.ResponseLogger()(nil, req)
}
