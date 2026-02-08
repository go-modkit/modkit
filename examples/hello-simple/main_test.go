// Package main provides entry point tests for the hello-simple example.
package main

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/module"
	"github.com/go-modkit/modkit/modkit/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResolver is a mock implementation of module.Resolver for testing.
type MockResolver struct {
	mock.Mock
}

// Get implements module.Resolver.
func (m *MockResolver) Get(token module.Token) (any, error) {
	args := m.Called(token)
	return args.Get(0), args.Error(1)
}

// TestAppModule_Definition tests the AppModule definition metadata.
func TestAppModule_Definition(t *testing.T) {
	m := NewAppModule("test message")
	def := m.Definition()

	assert.Equal(t, "app", def.Name)
	assert.Len(t, def.Providers, 2)
	assert.Len(t, def.Controllers, 1)
}

func TestGreetingController_Build(t *testing.T) {
	m := NewAppModule("test message")
	def := m.Definition()
	controllerDef := def.Controllers[0]

	t.Run("success", func(t *testing.T) {
		r := new(MockResolver)
		r.On("Get", TokenGreeting).Return("hello", nil)
		r.On("Get", TokenCounter).Return(&Counter{}, nil)

		val, err := controllerDef.Build(r)
		assert.NoError(t, err)
		assert.NotNil(t, val)
		assert.IsType(t, &GreetingController{}, val)
	})

	t.Run("missing greeting", func(t *testing.T) {
		r := new(MockResolver)
		r.On("Get", TokenGreeting).Return(nil, errors.New("missing"))

		val, err := controllerDef.Build(r)
		assert.Error(t, err)
		assert.Nil(t, val)
	})

	t.Run("missing counter", func(t *testing.T) {
		r := new(MockResolver)
		r.On("Get", TokenGreeting).Return("hello", nil)
		r.On("Get", TokenCounter).Return(nil, errors.New("missing"))

		val, err := controllerDef.Build(r)
		assert.Error(t, err)
		assert.Nil(t, val)
	})

	t.Run("type mismatch greeting", func(t *testing.T) {
		r := new(MockResolver)
		r.On("Get", TokenGreeting).Return(123, nil)

		val, err := controllerDef.Build(r)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolved to int, expected string")
		assert.Nil(t, val)
	})
}

func TestAppModule_Providers(t *testing.T) {
	m := NewAppModule("test message")
	def := m.Definition()

	t.Run("greeting message success", func(t *testing.T) {
		p := def.Providers[0]
		val, err := p.Build(nil)
		assert.NoError(t, err)
		assert.Equal(t, "test message", val)
	})

	t.Run("counter success", func(t *testing.T) {
		p := def.Providers[1]
		val, err := p.Build(nil)
		assert.NoError(t, err)
		assert.IsType(t, &Counter{}, val)
	})
}

func TestAppModule_TestKitOverrideGreeting(t *testing.T) {
	h := testkit.New(t,
		NewAppModule("real"),
		testkit.WithOverrides(testkit.OverrideValue(TokenGreeting, "fake")),
	)

	controller := testkit.Controller[*GreetingController](t, h, "app", "GreetingController")
	assert.Equal(t, "fake", controller.message)
}
