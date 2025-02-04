package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/livepeer/catalyst-api/config"
	"github.com/livepeer/catalyst-api/errors"
	"github.com/livepeer/catalyst-api/log"
	"github.com/prometheus/client_golang/prometheus"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Flush() {
	flusher := rw.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

func LogAndMetrics(metric *prometheus.SummaryVec) func(httprouter.Handle) httprouter.Handle {
	return logRequest(metric)
}

func LogRequest() func(httprouter.Handle) httprouter.Handle {
	return logRequest(nil)
}

func logRequest(metric *prometheus.SummaryVec) func(httprouter.Handle) httprouter.Handle {
	return func(next httprouter.Handle) httprouter.Handle {
		fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			defer func() {
				if err := recover(); err != nil {
					errors.WriteHTTPInternalServerError(wrapped, "Internal Server Error", nil)
					log.LogNoRequestID("returning HTTP 500", "err", err, "trace", debug.Stack())
				}
			}()

			next(wrapped, r, ps)
			duration := time.Since(start)
			log.LogNoRequestID("received HTTP request",
				"remote", r.RemoteAddr,
				"proto", r.Proto,
				"method", r.Method,
				"uri", r.URL.RequestURI(),
				"duration", duration,
				"status", wrapped.status,
			)

			if metric != nil {
				success := wrapped.status < 400
				metric.
					WithLabelValues(strconv.FormatBool(success), fmt.Sprint(wrapped.status), config.Version).
					Observe(duration.Seconds())
			}
		}

		return fn
	}
}
