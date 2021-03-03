package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	loggerTests = []struct {
		desc          string
		hasOutputFunc bool
		msg           string
		expected      string
	}{
		{
			desc:          "output func is nil",
			hasOutputFunc: false,
			msg:           "test logger",
			expected:      "",
		},
		{
			desc:          "output func is not nil",
			hasOutputFunc: true,
			msg:           "test logger",
			expected:      "test logger",
		},
	}
)

func TestWarn(t *testing.T) {
	t.Parallel()
	for _, lt := range loggerTests {
		t.Run(lt.desc, func(t *testing.T) {
			var buf bytes.Buffer
			var l *Logger
			if lt.hasOutputFunc {
				l = NewLogger(bufferedOutputFunc(t, &buf), nil)
			} else {
				l = NewLogger(nil, nil)
			}
			l.Warn(lt.msg)
			assert.Equal(t, lt.expected, buf.String())
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	for _, lt := range loggerTests {
		t.Run(lt.desc, func(t *testing.T) {
			var buf bytes.Buffer
			var l *Logger
			if lt.hasOutputFunc {
				l = NewLogger(nil, bufferedOutputFunc(t, &buf))
			} else {
				l = NewLogger(nil, nil)
			}
			l.Error(lt.msg)
			assert.Equal(t, lt.expected, buf.String())
		})
	}
}

var (
	loggerfTests = []struct {
		desc          string
		hasOutputFunc bool
		format        string
		args          []interface{}
		expected      string
	}{
		{
			desc:          "output func is nil",
			hasOutputFunc: false,
			format:        "test logger %v",
			args:          []interface{}{"arg"},
			expected:      "",
		},
		{
			desc:          "output func is not nil",
			hasOutputFunc: true,
			format:        "test logger %v",
			args:          []interface{}{"arg"},
			expected:      "test logger arg",
		},
	}
)

func TestWarnf(t *testing.T) {
	t.Parallel()
	for _, lt := range loggerfTests {
		t.Run(lt.desc, func(t *testing.T) {
			var buf bytes.Buffer
			var l *Logger
			if lt.hasOutputFunc {
				l = NewLogger(bufferedOutputFunc(t, &buf), nil)
			} else {
				l = NewLogger(nil, nil)
			}
			l.Warnf(lt.format, lt.args...)
			assert.Equal(t, lt.expected, buf.String())
		})
	}
}

func TestErrorf(t *testing.T) {
	t.Parallel()
	for _, lt := range loggerfTests {
		t.Run(lt.desc, func(t *testing.T) {
			var buf bytes.Buffer
			var l *Logger
			if lt.hasOutputFunc {
				l = NewLogger(nil, bufferedOutputFunc(t, &buf))
			} else {
				l = NewLogger(nil, nil)
			}
			l.Errorf(lt.format, lt.args...)
			assert.Equal(t, lt.expected, buf.String())
		})
	}
}

func bufferedOutputFunc(t *testing.T, buf *bytes.Buffer) OutputFunc {
	t.Helper()
	return func(msg string) {
		buf.WriteString(msg)
	}
}
