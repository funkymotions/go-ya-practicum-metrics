// Package handler provides HTTP handlers for working with application metrics.
//
// The package defines the metricHandler type and related HTTP endpoints for:
//   - updating metrics via plain-text and JSON APIs
//   - retrieving single metrics in plain-text and JSON formats
//   - rendering all metrics as HTML
//   - performing health checks (ping endpoint)
//
// These handlers are intended to be registered on a chi.Mux router by
// calling the Register method on metricHandler.
package handler
