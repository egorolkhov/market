package logger

import (
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

func PostLogger(h http.HandlerFunc) http.HandlerFunc {
	Foo := func(w http.ResponseWriter, r *http.Request) {
		Time := time.Now()
		Data := &responseData{0, 0}
		lw := LogResponseWriter{w, Data}
		h(&lw, r)
		Duration := time.Since(Time)
		Log.Info(
			"INFO",
			zap.String("method", r.Method),
			zap.String("status code", strconv.Itoa(Data.StatusCode)),
			zap.String("URI", r.RequestURI),
			zap.String("size", strconv.Itoa(Data.Size)),
			zap.String("time", Duration.String()),
		)
	}
	return Foo
}
