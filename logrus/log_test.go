package httplogrus

import (
	"io/ioutil"
	"testing"

	"github.com/Bplotka/go-httplog"
	"github.com/Sirupsen/logrus"
)

func TestToHTTPFieldLogger_CanBeUsed(t *testing.T) {
	l := logrus.New()
	l.Out = ioutil.Discard

	logger := ToHTTPFieldLoggerInfo(l)
	logger.WithFields(httplog.Fields{}).Log("something.")
}
