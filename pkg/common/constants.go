// Package common provides shared constants and utilities for the lsweb application.
package common

import "time"

// HTTP constants
const (
	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// UserAgent is the user agent string used for HTTP requests
	UserAgent = "lsweb/1.0"

	// MaxContentSize is the maximum content size for HTTP responses
	MaxContentSize = 10 * 1024 * 1024 // 10MB limit
)

// Version is the current version of the lsweb application
const Version = "1.0.0"
