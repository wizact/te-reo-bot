// Package logger provides structured logging capabilities with stack trace support
// for the Te Reo Bot application.
//
// This package implements a centralized logging system that captures and formats
// stack traces consistently across all components while maintaining security by
// ensuring stack traces are never exposed in HTTP responses to clients.
//
// The logger supports configurable stack trace logging through environment
// variables and provides structured logging with contextual fields.
package logger
