package module_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-modkit/modkit/modkit/module"
)

type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) Get(token module.Token) (any, error) {
	args := m.Called(token)
	return args.Get(0), args.Error(1)
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resolver := new(MockResolver)
		expected := "hello"
		token := module.Token("test")

		resolver.On("Get", token).Return(expected, nil)

		val, err := module.Get[string](resolver, token)
		assert.NoError(t, err)
		assert.Equal(t, expected, val)
	})

	t.Run("provider error", func(t *testing.T) {
		resolver := new(MockResolver)
		token := module.Token("test")
		expectedErr := errors.New("fail")

		resolver.On("Get", token).Return(nil, expectedErr)

		_, err := module.Get[string](resolver, token)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("type mismatch", func(t *testing.T) {
		resolver := new(MockResolver)
		token := module.Token("test")

		// Return int, expect string
		resolver.On("Get", token).Return(123, nil)

		_, err := module.Get[string](resolver, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"test\" resolved to int, expected string")
	})
}
