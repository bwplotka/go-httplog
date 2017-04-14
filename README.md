# go-httplog

Robust, smart logger for Golang http request/response.

It provides way to log in details every request and response with chosen fields. It
requires any structured logger that fits under `httplog.FieldLogger` interface.

It comes with useful integration with [logrus]("github.com/Sirupsen/logrus"), but it can be extend to use any logger.

It fits into standard `net/http` middleware (`http.Handler`) pattern, but comes also with [echo]("github.com/labstack/echo") integration.

It comes with bunch of configurable fields. Not exhausting list of these:

```
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
	
	StatusField       = ResponseField("res_status")
	BytesOutField     = ResponseField("res_bytes_out")
	ResTimeField      = ResponseField("res_time")
	ContentTypeField  = ResponseField("res_content_type")
	LocationField     = ResponseField("res_location")
	LocationArgsField = ResponseField("res_location_args")
	LocationHostField = ResponseField("res_location_host")
```

## Example:

```go
package main

import (
    "net/http"
    
    "github.com/Bplotka/go-httplog"
    "github.com/Bplotka/go-httplog/logrus"
    "github.com/Sirupsen/logrus"
)

func main() {
    l := logrus.New()
    httpLogger := httplog.New(
        httplogrus.ToHTTPFieldLoggerDebug(l), // or ToHTTPFieldLoggerInfo if you want these logs to be in Info level.
        httplog.DefaultReqResConfig(), // or httplog.DefaultResponseOnlyConfig() for only log line per response. 
    )
    
    srv := http.Server{
        Handler: httpLogger.ResponseMiddleware()(
            http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                httpLogger.RequestLogger()(w, r)
            
                // Your handler here...
                l.Info("Inside!")
            
            }),
        ),
    }
    
    err := srv.Serve(...)
    if err != nil {
        // handle err.
    }
}
```

Example effect: 

* With `DefaultReqResConfig` on any request you will get:
```
Debug[0029] Received HTTP request                         req_bytes_in=0 req_host="127.0.0.1:<some-port>" req_method=GET req_path="...." req_remote_ip=127.0.0.1 req_time="2017-04-14T17:20:07+01:00" <any other field that will filled and configured>
Inside!
Debug[0029] Responding to HTTP request                    res_bytes_out=769 res_content_type="application/json" res_status=200 res_time="2017-04-14T17:20:07+01:00" <any other field that will filled and configured>
```

For redirection the log can be even more useful:
```
Debug[0029] Received HTTP request                         req_bytes_in=0 req_host="127.0.0.1:<some-port>" req_method=GET req_path="...." req_remote_ip=127.0.0.1 req_time="2017-04-14T17:20:07+01:00" <any other field that will filled and configured>
Inside!
Debug[0029] Redirecting HTTP request                      res_bytes_out=0 res_location_args="code=...&state=..." res_location_host="...." res_status=303 res_time="2017-04-14T17:20:07+01:00" <any other field that will filled and configured>" 
```

## Integrations:

### Web frameworks
* net/http ;>
* [echo](echo/middleware.go)

### Structured Loggers
* [logrus](logrus/log.go)
