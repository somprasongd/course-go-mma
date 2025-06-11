package registry

import "fmt"

func ResolveAs[T any](r ServiceRegistry, key ServiceKey) (T, error) {
	var zero T
	svc, err := r.Resolve(key)
	if err != nil {
		return zero, err
	}
	typedSvc, ok := svc.(T)
	if !ok {
		return zero, fmt.Errorf("service registered under key %s does not implement the expected type", key)
	}
	return typedSvc, nil
}
