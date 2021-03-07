package log

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := Loggers{
		debugLogger: &bufLogger{&buf},
	}
	msg := "message"
	l.Debug(msg)
	assert.Equal(t, msg, buf.String())
}

func TestDebugf(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := Loggers{
		debugLogger: &bufLogger{&buf},
	}
	l.Debugf("message: %d", 1)
	expected := "message: 1"
	assert.Equal(t, expected, buf.String())
}

func TestError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := Loggers{
		errorLogger: &bufLogger{&buf},
	}
	msg := "message"
	l.Error(msg)
	assert.Equal(t, msg, buf.String())
}

func TestErrorf(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := Loggers{
		errorLogger: &bufLogger{&buf},
	}
	l.Errorf("message: %d", 1)
	expected := "message: 1"
	assert.Equal(t, expected, buf.String())
}

type bufLogger struct {
	b *bytes.Buffer
}

func (l *bufLogger) Print(values ...interface{}) {
	l.b.WriteString(fmt.Sprint(values...))
}

func (l *bufLogger) Printf(format string, values ...interface{}) {
	l.b.WriteString(fmt.Sprintf(format, values...))
}
