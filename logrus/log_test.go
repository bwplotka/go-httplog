package httplogrus

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestToHTTPFieldLogger_CanBeUsed(t *testing.T) {
	l := logrus.New()
	l.Out = ioutil.Discard

	logger := ToHTTPFieldLogger(l)
	logger.WithError(errors.New("error")).Debug("something.")
}
