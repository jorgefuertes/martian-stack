package httpconst

import "strings"

type Method string

const (
	MethodGet     Method = "GET"
	MethodHead    Method = "HEAD"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH" // RFC 5789
	MethodDelete  Method = "DELETE"
	MethodConnect Method = "CONNECT"
	MethodOptions Method = "OPTIONS"
	MethodTrace   Method = "TRACE"
	MethodAny     Method = "ANY"
)

func (m Method) String() string {
	return string(m)
}

func IsValidMethod(m Method) bool {
	return strings.Contains("GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE,ANY", m.String())
}

func IsMethodAny(m Method) bool {
	return m == MethodAny
}
