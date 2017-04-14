package httplog

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

//go:generate mockery -name FieldLogger -case underscore -inpkg

func testRequest(url string) (*http.Request, error) {
	return http.NewRequest(
		"GET",
		url+`/some_endpoint?arg1="arg1value"&arg2="arg2value"`,
		nil,
	)
}

func TestLogger_HTTPHandler_EmptyConfig(t *testing.T) {
	mLogger := new(MockFieldLogger)
	l := New(mLogger, Config{})

	req, err := testRequest("")
	require.NoError(t, err)
	l.RequestLogger()(nil, req)

	mLogger.AssertExpectations(t)
}

func TestLogger_HTTPHandler_RequestFieldsLogged(t *testing.T) {
	mLogger := new(MockFieldLogger)
	now := time.Now()
	mLogger.On("WithFields", Fields{
		"req_bytes_in": "0",
		"req_args":     "arg1=...&arg2=...",
		"req_path":     "/some_endpoint",
		"req_time":     now.Format(time.RFC3339),
		"req_method":   "GET",
	}).Return(mLogger)
	mLogger.On("Log", []interface{}{"Received HTTP request"}).Once()

	l := New(mLogger, DefaultReqResConfig())
	l.timeNow = func() time.Time {
		return now
	}

	req, err := testRequest("")
	require.NoError(t, err)
	l.RequestLogger()(nil, req)

	mLogger.AssertExpectations(t)
}

func TestLogger_HTTPHandler_ResponseFieldsLogged(t *testing.T) {
	mLogger := new(MockFieldLogger)
	now := time.Now()
	mLogger.On("WithFields", Fields{
		"res_bytes_out":    "39",
		"res_content_type": "application/json",
		"res_time":         now.Format(time.RFC3339),
		"res_status":       "200",
	}).Return(mLogger)
	mLogger.On("Log", []interface{}{"Responding to HTTP request"}).Once()

	l := New(mLogger, DefaultReqResConfig())
	l.timeNow = func() time.Time {
		return now
	}

	srv := httptest.NewServer(l.ResponseMiddleware()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ioutil.ReadAll(r.Body)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")

			data := struct {
				FieldA string `json:"field_a"`
				FieldB bool   `json:"field_b"`
			}{
				FieldA: "something",
				FieldB: true,
			}
			encoder := json.NewEncoder(w)
			err := encoder.Encode(data)
			require.NoError(t, err)
		}),
	))

	defer srv.Close()

	req, err := testRequest(srv.URL)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	ioutil.ReadAll(res.Body)

	mLogger.AssertExpectations(t)
}

func TestLogger_HTTPHandler_RedirectFieldsLogged(t *testing.T) {
	mLogger := new(MockFieldLogger)
	now := time.Now()
	mLogger.On("WithFields", Fields{
		"res_status":        "302",
		"res_bytes_out":     "0",
		"res_time":          now.Format(time.RFC3339),
		"res_location_args": "arg1=...&arg2=...",
		"res_location_host": "http://localhost/wrong_endpoint",
	}).Return(mLogger)
	mLogger.On("Log", []interface{}{"Redirecting HTTP request"}).Once()

	l := New(mLogger, DefaultReqResConfig())
	l.timeNow = func() time.Time {
		return now
	}

	srv := httptest.NewServer(l.ResponseMiddleware()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ioutil.ReadAll(r.Body)

			u := `http://localhost/wrong_endpoint?arg1="arg1value"&arg2="arg2value"`
			http.Redirect(w, r, u, 302)
		}),
	))

	defer srv.Close()

	req, err := testRequest(srv.URL)
	require.NoError(t, err)

	_, err = http.DefaultClient.Do(req)
	require.Error(t, err)

	mLogger.AssertExpectations(t)
}

func TestFormatComactArgs(t *testing.T) {
	u := `http://localhost/wrong_endpoint?arg1="arg1value"&arg2="arg2value"`
	url, err := url.Parse(u)
	require.NoError(t, err)

	assert.Equal(t, "arg1=...&arg2=...", formatCompactArgs(url.RawQuery))

}
