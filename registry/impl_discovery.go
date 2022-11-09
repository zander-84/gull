package registry

import (
	"context"
	"github.com/zander-84/gull/contrib/lb"
	"time"
)

type ServiceDiscovery struct {
	d        Discovery
	lb       lb.Balancer
	listener lb.Listener
	watcher  Watcher
}

func NewServiceDiscovery(serviceName string, d Discovery, p lb.Policy) (*ServiceDiscovery, error) {
	var err error
	sd := new(ServiceDiscovery)
	sd.d = d
	sd.listener = lb.NewListener(serviceName)
	sd.lb = lb.NewBalancer(sd.listener, p, false)
	sd.watcher, err = sd.d.Watch(context.Background(), serviceName)
	if err != nil {
		return nil, err
	}
	_ = sd.set()

	go func() {
		for {
			if err := sd.set(); err != nil {
				time.Sleep(time.Second)
				continue
			}
		}
	}()
	return sd, nil
}

func (sd *ServiceDiscovery) set() error {
	serviceInstances, err := sd.watcher.Next()
	if err != nil {
		return err
	}
	in := make(map[any]int, 0)
	for _, serviceInstance := range serviceInstances {
		in[serviceInstance] = serviceInstance.Weight
	}
	sd.listener.Set(in)
	return nil
}
func (sd *ServiceDiscovery) GetServiceInstance() (ServiceInstance, error) {
	outAny, err := sd.lb.Next()
	if err != nil {
		return ServiceInstance{}, err
	}
	out := outAny.(*ServiceInstance)
	return *out, nil
}
