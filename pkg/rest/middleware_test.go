package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func mw(name string, order *[]string, block bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, "pre:"+name)
			if block {
				http.Error(w, "blocked by "+name, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			*order = append(*order, "post:"+name)
		})
	}
}

func Test_ApplyMiddlewaresChain(t *testing.T) {
	var order []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusNoContent)
	})

	chained := applyMiddlewares(handler, []func(http.Handler) http.Handler{
		mw("a", &order, false),
		mw("b", &order, false),
	})

	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, []string{"pre:a", "pre:b", "handler", "post:b", "post:a"}, order)
}

func Test_ApplyMiddlewaresBreak(t *testing.T) {
	var order []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusNoContent)
	})

	chained := applyMiddlewares(handler, []func(http.Handler) http.Handler{
		mw("a", &order, true),
		mw("b", &order, false),
	})

	rec := httptest.NewRecorder()
	chained.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "blocked by a")
	require.Equal(t, []string{"pre:a"}, order)
}
