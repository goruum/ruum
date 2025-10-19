// Package events provides event emitter functionality similar to NestJS EventEmitter.
package events

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Event represents an event
type Event struct {
	Name    string
	Payload interface{}
	Context context.Context
}

// Listener is a function that handles an event
type Listener func(event Event) error

// AsyncListener is an async function that handles an event
type AsyncListener func(event Event)

// EventEmitter manages event listeners and dispatches events
type EventEmitter struct {
	listeners      map[string][]Listener
	asyncListeners map[string][]AsyncListener
	mu             sync.RWMutex
	maxListeners   int
	wildcard       bool
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		listeners:      make(map[string][]Listener),
		asyncListeners: make(map[string][]AsyncListener),
		maxListeners:   10,
		wildcard:       false,
	}
}

// SetMaxListeners sets the maximum number of listeners per event
func (e *EventEmitter) SetMaxListeners(max int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = max
}

// On registers a synchronous listener for an event
func (e *EventEmitter) On(eventName string, listener Listener) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.listeners[eventName]) >= e.maxListeners {
		return fmt.Errorf("maximum listeners reached for event: %s", eventName)
	}

	e.listeners[eventName] = append(e.listeners[eventName], listener)
	return nil
}

// Once registers a listener that will only be called once
func (e *EventEmitter) Once(eventName string, listener Listener) error {
	wrapper := func(event Event) error {
		e.RemoveListener(eventName, listener)
		return listener(event)
	}
	return e.On(eventName, wrapper)
}

// OnAsync registers an asynchronous listener for an event
func (e *EventEmitter) OnAsync(eventName string, listener AsyncListener) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.asyncListeners[eventName]) >= e.maxListeners {
		return fmt.Errorf("maximum async listeners reached for event: %s", eventName)
	}

	e.asyncListeners[eventName] = append(e.asyncListeners[eventName], listener)
	return nil
}

// Emit dispatches an event synchronously to all registered listeners
func (e *EventEmitter) Emit(eventName string, payload interface{}) error {
	return e.EmitWithContext(context.Background(), eventName, payload)
}

// EmitWithContext dispatches an event with context
func (e *EventEmitter) EmitWithContext(ctx context.Context, eventName string, payload interface{}) error {
	event := Event{
		Name:    eventName,
		Payload: payload,
		Context: ctx,
	}

	e.mu.RLock()
	syncListeners := make([]Listener, len(e.listeners[eventName]))
	copy(syncListeners, e.listeners[eventName])

	asyncListeners := make([]AsyncListener, len(e.asyncListeners[eventName]))
	copy(asyncListeners, e.asyncListeners[eventName])
	e.mu.RUnlock()

	// Execute sync listeners
	for _, listener := range syncListeners {
		if err := listener(event); err != nil {
			return fmt.Errorf("listener error for event %s: %w", eventName, err)
		}
	}

	// Execute async listeners in goroutines
	for _, listener := range asyncListeners {
		go listener(event)
	}

	return nil
}

// EmitAsync dispatches an event asynchronously
func (e *EventEmitter) EmitAsync(eventName string, payload interface{}) {
	go func() {
		_ = e.Emit(eventName, payload)
	}()
}

// RemoveListener removes a specific listener
func (e *EventEmitter) RemoveListener(eventName string, listener Listener) {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners := e.listeners[eventName]
	for i, l := range listeners {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			e.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// RemoveAllListeners removes all listeners for an event
func (e *EventEmitter) RemoveAllListeners(eventName string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.listeners, eventName)
	delete(e.asyncListeners, eventName)
}

// ListenerCount returns the number of listeners for an event
func (e *EventEmitter) ListenerCount(eventName string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners[eventName]) + len(e.asyncListeners[eventName])
}

// EventNames returns all event names that have listeners
func (e *EventEmitter) EventNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	names := make(map[string]bool)
	for name := range e.listeners {
		names[name] = true
	}
	for name := range e.asyncListeners {
		names[name] = true
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}

	return result
}

// EventBus is a global event bus
type EventBus struct {
	emitters map[string]*EventEmitter
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		emitters: make(map[string]*EventEmitter),
	}
}

// GetEmitter gets or creates an emitter for a namespace
func (eb *EventBus) GetEmitter(namespace string) *EventEmitter {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if emitter, exists := eb.emitters[namespace]; exists {
		return emitter
	}

	emitter := NewEventEmitter()
	eb.emitters[namespace] = emitter
	return emitter
}

// Emit emits an event on a specific namespace
func (eb *EventBus) Emit(namespace, eventName string, payload interface{}) error {
	emitter := eb.GetEmitter(namespace)
	return emitter.Emit(eventName, payload)
}

// On registers a listener on a specific namespace
func (eb *EventBus) On(namespace, eventName string, listener Listener) error {
	emitter := eb.GetEmitter(namespace)
	return emitter.On(eventName, listener)
}

// OnAsync registers an async listener on a specific namespace
func (eb *EventBus) OnAsync(namespace, eventName string, listener AsyncListener) error {
	emitter := eb.GetEmitter(namespace)
	return emitter.OnAsync(eventName, listener)
}

// DefaultEventBus is the default global event bus
var DefaultEventBus = NewEventBus()

// Global convenience functions

// On registers a listener on the default event bus
func On(eventName string, listener Listener) error {
	return DefaultEventBus.GetEmitter("default").On(eventName, listener)
}

// OnAsync registers an async listener on the default event bus
func OnAsync(eventName string, listener AsyncListener) error {
	return DefaultEventBus.GetEmitter("default").OnAsync(eventName, listener)
}

// Emit emits an event on the default event bus
func Emit(eventName string, payload interface{}) error {
	return DefaultEventBus.GetEmitter("default").Emit(eventName, payload)
}

// EmitAsync emits an event asynchronously on the default event bus
func EmitAsync(eventName string, payload interface{}) {
	DefaultEventBus.GetEmitter("default").EmitAsync(eventName, payload)
}
