package registry

import (
	"context"
	"fmt"
	"sync"
)

type Registry struct {
	lk      sync.Mutex
	Service map[string]*ServiceInstance
	Engine  Registrar
}

func (r *Registry) Register(ctx context.Context, service *ServiceInstance) error {
	if service == nil || service.ID == "" {
		return fmt.Errorf("no service id")
	}
	r.lk.Lock()
	defer r.lk.Unlock()
	if r.Service == nil {
		r.Service = map[string]*ServiceInstance{}
	}
	r.Service[service.ID] = service
	for _, v := range r.Service {
		if err := r.Engine.Register(ctx, v); err != nil {
			return err
		}
	}
	return nil
}

// Deregister the registration.
func (r *Registry) Deregister(ctx context.Context, service *ServiceInstance) error {
	r.lk.Lock()
	defer r.lk.Unlock()
	if r.Service[service.ID] == nil {
		return fmt.Errorf("deregister service not found")
	}

	for _, v := range r.Service {
		if err := r.Engine.Deregister(ctx, v); err != nil {
			return err
		}
	}
	return nil
}
