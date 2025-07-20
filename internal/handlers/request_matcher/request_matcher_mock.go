package request_matcher

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockRequestMatcher struct {
	mock.Mock
}

func (m *MockRequestMatcher) Match(r *http.Request, method, pathPattern string) (bool, map[string]string, error) {
	args := m.Called(r, method, pathPattern)
	return args.Bool(0), args.Get(1).(map[string]string), args.Error(2)
}

func (m *MockRequestMatcher) ExtractPathParameters(requestPath, pathPattern string) (map[string]string, error) {
	args := m.Called(requestPath, pathPattern)
	return args.Get(0).(map[string]string), args.Error(1)
}
