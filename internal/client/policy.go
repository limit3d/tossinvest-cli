package client

type EndpointPolicy struct {
	allowed map[string]struct{}
}

func NewReadOnlyPolicy(paths []string) EndpointPolicy {
	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[path] = struct{}{}
	}

	return EndpointPolicy{allowed: allowed}
}

func (p EndpointPolicy) IsAllowed(path string) bool {
	_, ok := p.allowed[path]
	return ok
}
