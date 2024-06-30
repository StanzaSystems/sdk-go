package handlers

import (
	"log"
	"net/http"
)

type middleware struct {
	guardName     string
	featureName   string
	priorityBoost int32
	defaultWeight float32
	tags          map[string]string
}

// NewHttpHandler wraps the passed handler and enriches it with a guard.
func NewHttpHandler(handler http.Handler, guardName string, featureName *string,
	priorityBoost *int32, defaultWeight *float32, tags *map[string]string) http.Handler {
	return NewMiddleware(guardName, featureName, priorityBoost, defaultWeight, tags)(handler)
}

// NewMiddleware returns a Guard middleware
// The handler returned by the middleware wraps a handler and provides a Guard
// for each request
func NewMiddleware(guardName string, featureName *string, priorityBoost *int32,
	defaultWeight *float32, tags *map[string]string) func(http.Handler) http.Handler {
	h := middleware{
		guardName:     guardName,
		featureName:   *featureName,
		priorityBoost: *priorityBoost,
		defaultWeight: *defaultWeight,
		tags:          *tags,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.serveHTTP(w, r, next)
		})
	}
}

func (h *middleware) serveHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	ctx := r.Context()

	handler, err := NewHandler(h.guardName, &h.featureName, &h.priorityBoost, &h.defaultWeight, &h.tags)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("Error creating handler: %v", err)
	}

	NewInboundHandler(h.guardName, &h.featureName, &h.priorityBoost, &h.defaultWeight, &h.tags)

	guard := handler.Guard(ctx, nil, nil)
	if guard.Error() != nil {
		guard.End(guard.Failure)
	}
	if guard.Allowed() {
		guard.End(guard.Success)
		next.ServeHTTP(w, r)
	} else {
		guard.End(guard.Failure)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	}
}
