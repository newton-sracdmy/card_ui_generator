package dto

import "context"

type RequestOptions struct {
	Context     context.Context
	Headers     map[string]string
	AuthToken   string
	CompanyID   string
	RequestID   string
	ContentType string
}

type FilePart struct {
	Field    string
	Filename string
	Data     []byte
}

type SendRequestParams struct {
	Method  string
	BaseURL string
	Path    string
	Body    any
	Dest    any
	Opts    RequestOptions
}

type SendMultipartRequestParams struct {
	Method string
	URL    string
	Fields map[string]string
	Files  []FilePart
	Dest   any
	Opts   RequestOptions
}

type DoJSONParams struct {
	Method string
	URL    string
	Body   any
	Opts   RequestOptions
}
