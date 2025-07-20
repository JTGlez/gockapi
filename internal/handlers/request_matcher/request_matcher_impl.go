package request_matcher

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type RequestMatcherImpl struct{}

func (rm *RequestMatcherImpl) Match(r *http.Request, method, pathPattern string) (bool, map[string]string, error) {
	if !strings.EqualFold(r.Method, method) {
		return false, nil, nil
	}

	pathParams, err := rm.ExtractPathParameters(r.URL.Path, pathPattern)
	if err != nil {
		return false, nil, err
	}

	if pathParams == nil {
		return r.URL.Path == pathPattern, nil, nil
	}

	return true, pathParams, nil
}

func (rm *RequestMatcherImpl) ExtractPathParameters(requestPath, pathPattern string) (map[string]string, error) {
	if !strings.Contains(pathPattern, "{") {
		if requestPath == pathPattern {
			return map[string]string{}, nil
		}

		return nil, nil
	}

	regexPattern := pathPattern
	paramNames := []string{}

	paramRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := paramRegex.FindAllStringSubmatch(pathPattern, -1)

	for _, match := range matches {
		paramName := match[1]
		paramNames = append(paramNames, paramName)
		regexPattern = strings.Replace(regexPattern, match[0], `([^/]+)`, 1)
	}

	regexPattern = "^" + regexPattern + "$"

	compiled, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid path pattern %s: %w", pathPattern, err)
	}

	matches = compiled.FindAllStringSubmatch(requestPath, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	pathParams := make(map[string]string)

	for i, paramName := range paramNames {
		if i+1 < len(matches[0]) {
			pathParams[paramName] = matches[0][i+1]
		}
	}

	return pathParams, nil
}
