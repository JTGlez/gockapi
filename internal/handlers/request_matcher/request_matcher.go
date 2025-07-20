package request_matcher

import "net/http"

type RequestMatcher interface {
	Match(r *http.Request, method, pathPattern string) (bool, map[string]string, error)
	ExtractPathParameters(requestPath, pathPattern string) (map[string]string, error)
}
