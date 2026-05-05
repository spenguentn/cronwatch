package notify

import (
	"context"
	"fmt"
	"sync"
)

// Route maps a named key to a Notifier.
type Route struct {
	Name     string
	Notifier Notifier
}

// RoutingNotifier dispatches messages to a specific notifier chosen by
// a RoutingFunc. If no route matches, the fallback notifier is used (if set).
type RoutingNotifier struct {
	mu       sync.RWMutex
	routes   map[string]Notifier
	fallback Notifier
	keyFn    func(Message) string
}

// NewRoutingNotifier creates a RoutingNotifier. keyFn extracts the routing key
// from a message (e.g. job name from Meta). routes maps keys to notifiers.
// fallback may be nil, in which case unmatched messages are silently dropped.
func NewRoutingNotifier(keyFn func(Message) string, fallback Notifier, routes ...Route) *RoutingNotifier {
	rn := &RoutingNotifier{
		routes:   make(map[string]Notifier, len(routes)),
		fallback: fallback,
		keyFn:    keyFn,
	}
	for _, r := range routes {
		rn.routes[r.Name] = r.Notifier
	}
	return rn
}

// AddRoute registers or replaces a route at runtime.
func (rn *RoutingNotifier) AddRoute(name string, n Notifier) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.routes[name] = n
}

// RemoveRoute deletes a named route.
func (rn *RoutingNotifier) RemoveRoute(name string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	delete(rn.routes, name)
}

// Notify resolves the route and forwards the message.
func (rn *RoutingNotifier) Notify(ctx context.Context, msg Message) error {
	if rn.keyFn == nil {
		return fmt.Errorf("routing: keyFn is nil")
	}
	key := rn.keyFn(msg)

	rn.mu.RLock()
	n, ok := rn.routes[key]
	rn.mu.RUnlock()

	if !ok {
		if rn.fallback != nil {
			return rn.fallback.Notify(ctx, msg)
		}
		return nil
	}
	return n.Notify(ctx, msg)
}
