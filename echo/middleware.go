package echolog

import (
	"github.com/Bplotka/go-httplog"
	"github.com/labstack/echo"
)

// RegisterMiddleware registers echo handler that will log request at the beginning and served response at the request end.
func RegisterMiddleware(logger httplog.FieldLogger, cfg httplog.Config) echo.MiddlewareFunc {
	l := httplog.New(logger, cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// Log specified RequestFields now.
			l.RequestHandler()(c.Response(), c.Request())

			// Wrap ResponseWriter under echo.Response to log specified ResponseFields and ResponseReqFields on
			// Response Write or Redirect.
			w := l.WrapResponse(c.Response().Writer, c.Request())
			c.Response().Writer = w
			return next(c)
		}
	}
}
