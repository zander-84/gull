package app

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/zander-84/gull/contrib/endpoint/http_router"
	"github.com/zander-84/gull/contrib/registry/etcd"
	"github.com/zander-84/gull/endpoint"
	"github.com/zander-84/gull/pbs"
	"github.com/zander-84/gull/registry"
	"github.com/zander-84/gull/think"
	"github.com/zander-84/gull/transport/grpc"
	"github.com/zander-84/gull/transport/http"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpc2 "google.golang.org/grpc"
	"net/url"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"
)

type mockRegistry struct {
	lk      sync.Mutex
	service map[string]*registry.ServiceInstance
}

func (r *mockRegistry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	if service == nil || service.ID == "" {
		return fmt.Errorf("no service id")
	}
	r.lk.Lock()
	defer r.lk.Unlock()
	r.service[service.ID] = service
	return nil
}

// Deregister the registration.
func (r *mockRegistry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	r.lk.Lock()
	defer r.lk.Unlock()
	if r.service[service.ID] == nil {
		return fmt.Errorf("deregister service not found")
	}
	delete(r.service, service.ID)
	return nil
}

func TestApp(t *testing.T) {
	resource := endpoint.NewRmc()
	resource = resource.Use(endpoint.OptionsErrorEncoder(endpoint.WrapError(map[endpoint.Protocol]func(ctx context.Context, err error){
		endpoint.Http: func(ctx context.Context, err error) {

		},
		endpoint.Grpc: func(ctx context.Context, err error) {

		},
	}))).Use(endpoint.OptionsRecoverEncoder(endpoint.WrapRecover(map[endpoint.Protocol]func(ctx context.Context){
		endpoint.Http: think.Recover,
		endpoint.Grpc: think.Recover,
	})))

	resource.Endpoint([]endpoint.Protocol{endpoint.Http, endpoint.Grpc}, endpoint.MethodGet, "/a", func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return "hello", err
	}, endpoint.WrapDecode(map[endpoint.Protocol]endpoint.HandlerFunc{
		endpoint.Http: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			return request, err
		},
		endpoint.Grpc: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			return request, nil
		},
	}), endpoint.WrapEncode(map[endpoint.Protocol]endpoint.HandlerFunc{
		endpoint.Http: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			httpCtx := ctx.(http.Context)
			err = httpCtx.String(200, "hello")
			return ctx, err
		},
		endpoint.Grpc: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			data := new(pbs.Response)
			data.AdminName = "zander"
			return data, err
		},
	}))

	//g := gin.New()
	g := httprouter.New()
	p := http_router.NewRouter(g)
	resource.Proxy(p.Endpoint, endpoint.Http)

	hs := http.NewServer("127.0.0.1:9009", http.ServerHandler(p))
	gs := grpc.NewServer("127.0.0.1:9003")
	pbs.RegisterAdminServiceServer(gs, &server{Rmc: resource})

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"172.16.86.160:2379"},
		DialTimeout: time.Second, DialOptions: []grpc2.DialOption{grpc2.WithBlock()},
	})

	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	app := New(
		Name("kratos"),
		Version("v1.0.0"),
		Server(hs, gs),
		Signal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
		RegistrarTimeout(time.Second*10),
		Registrar(&registry.Registry{Engine: etcd.New(client), Service: map[string]*registry.ServiceInstance{
			"kratos-1": {
				ID:        "kratos-1",
				Name:      "kratos-1",
				Version:   "v1.0.0",
				Endpoints: []string{"http://127.0.0.1:9009"},
			},
			"kratos-2": {
				ID:        "kratos-2",
				Name:      "kratos-2",
				Version:   "v1.0.0",
				Endpoints: []string{"grpc://127.0.0.1:9003"},
			},
		}}),
		//Registrar(&mockRegistry{service: map[string]*registry.ServiceInstance{}}),
	)
	//time.AfterFunc(time.Second, func() {
	//	_ = app.Stop()
	//})
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}
}

type server struct {
	Rmc endpoint.Rmc
	pbs.UnimplementedAdminServiceServer
}

func (s server) Info(ctx context.Context, in *pbs.Request) (*pbs.Response, error) {
	h := s.Rmc.MustGetEndpoint(endpoint.MethodGet, "/a")

	endpointCtxVal := endpoint.NewCtxVal()
	endpointCtxVal.SetProtocol(endpoint.Grpc)
	ctx = endpoint.WithContext(ctx, endpointCtxVal)

	data, err := h(grpc.NewGrpcContext(ctx), in)
	data2 := data.(*pbs.Response)
	return data2, err
	//return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}

func TestApp_ID(t *testing.T) {
	v := "123"
	o := New(ID(v))
	if !reflect.DeepEqual(v, o.ID()) {
		t.Fatalf("o.ID():%s is not equal to v:%s", o.ID(), v)
	}
}

func TestApp_Name(t *testing.T) {
	v := "123"
	o := New(Name(v))
	if !reflect.DeepEqual(v, o.Name()) {
		t.Fatalf("o.Name():%s is not equal to v:%s", o.Name(), v)
	}
}

func TestApp_Version(t *testing.T) {
	v := "123"
	o := New(Version(v))
	if !reflect.DeepEqual(v, o.Version()) {
		t.Fatalf("o.Version():%s is not equal to v:%s", o.Version(), v)
	}
}

func TestApp_Metadata(t *testing.T) {
	v := map[string]string{
		"a": "1",
		"b": "2",
	}
	o := New(Metadata(v))
	if !reflect.DeepEqual(v, o.Metadata()) {
		t.Fatalf("o.Metadata():%s is not equal to v:%s", o.Metadata(), v)
	}
}

func TestApp_Endpoint(t *testing.T) {
	v := []string{"https://go-kratos.dev", "localhost"}
	var endpoints []*url.URL
	for _, urlStr := range v {
		if endpoint, err := url.Parse(urlStr); err != nil {
			t.Errorf("invalid endpoint:%v", urlStr)
		} else {
			endpoints = append(endpoints, endpoint)
		}
	}
	o := New(Endpoint(endpoints...))
	if instance, err := o.buildInstance(); err != nil {
		t.Error("build instance failed")
	} else {
		o.instance = instance
	}
	if !reflect.DeepEqual(o.Endpoint(), v) {
		t.Errorf("Endpoint() = %v, want %v", o.Endpoint(), v)
	}
}

func TestApp_buildInstance(t *testing.T) {
	want := struct {
		id        string
		name      string
		version   string
		metadata  map[string]string
		endpoints []string
	}{
		id:      "1",
		name:    "kratos",
		version: "v1.0.0",
		metadata: map[string]string{
			"a": "1",
			"b": "2",
		},
		endpoints: []string{"https://go-kratos.dev", "localhost"},
	}
	var endpoints []*url.URL
	for _, urlStr := range want.endpoints {
		if endpoint, err := url.Parse(urlStr); err != nil {
			t.Errorf("invalid endpoint:%v", urlStr)
		} else {
			endpoints = append(endpoints, endpoint)
		}
	}
	app := New(
		ID(want.id),
		Name(want.name),
		Version(want.version),
		Metadata(want.metadata),
		Endpoint(endpoints...),
	)
	if got, err := app.buildInstance(); err != nil {
		t.Error("build got failed")
	} else {
		if got.ID != want.id {
			t.Errorf("ID() = %v, want %v", got.ID, want.id)
		}
		if got.Name != want.name {
			t.Errorf("Name() = %v, want %v", got.Name, want.name)
		}
		if got.Version != want.version {
			t.Errorf("Version() = %v, want %v", got.Version, want.version)
		}
		if !reflect.DeepEqual(got.Endpoints, want.endpoints) {
			t.Errorf("Endpoint() = %v, want %v", got.Endpoints, want.endpoints)
		}
		if !reflect.DeepEqual(got.Metadata, want.metadata) {
			t.Errorf("Metadata() = %v, want %v", got.Metadata, want.metadata)
		}
	}
}

func TestApp_Context(t *testing.T) {
	type fields struct {
		id       string
		version  string
		name     string
		instance *registry.ServiceInstance
		metadata map[string]string
		want     struct {
			id       string
			version  string
			name     string
			endpoint []string
			metadata map[string]string
		}
	}
	tests := []fields{
		{
			id:       "1",
			name:     "kratos-v1",
			instance: &registry.ServiceInstance{Endpoints: []string{"https://go-kratos.dev", "localhost"}},
			metadata: map[string]string{},
			version:  "v1",
			want: struct {
				id       string
				version  string
				name     string
				endpoint []string
				metadata map[string]string
			}{
				id: "1", version: "v1", name: "kratos-v1", endpoint: []string{"https://go-kratos.dev", "localhost"},
				metadata: map[string]string{},
			},
		},
		{
			id:       "2",
			name:     "kratos-v2",
			instance: &registry.ServiceInstance{Endpoints: []string{"test"}},
			metadata: map[string]string{"kratos": "https://github.com/go-kratos/kratos"},
			version:  "v2",
			want: struct {
				id       string
				version  string
				name     string
				endpoint []string
				metadata map[string]string
			}{
				id: "2", version: "v2", name: "kratos-v2", endpoint: []string{"test"},
				metadata: map[string]string{"kratos": "https://github.com/go-kratos/kratos"},
			},
		},
		{
			id:       "3",
			name:     "kratos-v3",
			instance: nil,
			metadata: make(map[string]string),
			version:  "v3",
			want: struct {
				id       string
				version  string
				name     string
				endpoint []string
				metadata map[string]string
			}{
				id: "3", version: "v3", name: "kratos-v3", endpoint: nil,
				metadata: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				opts:     options{id: tt.id, name: tt.name, metadata: tt.metadata, version: tt.version},
				ctx:      context.Background(),
				cancel:   nil,
				instance: tt.instance,
			}

			ctx := NewContext(context.Background(), a)

			if got, ok := FromContext(ctx); ok {
				if got.ID() != tt.want.id {
					t.Errorf("ID() = %v, want %v", got.ID(), tt.want.id)
				}
				if got.Name() != tt.want.name {
					t.Errorf("Name() = %v, want %v", got.Name(), tt.want.name)
				}
				if got.Version() != tt.want.version {
					t.Errorf("Version() = %v, want %v", got.Version(), tt.want.version)
				}
				if !reflect.DeepEqual(got.Endpoint(), tt.want.endpoint) {
					t.Errorf("Endpoint() = %v, want %v", got.Endpoint(), tt.want.endpoint)
				}
				if !reflect.DeepEqual(got.Metadata(), tt.want.metadata) {
					t.Errorf("Metadata() = %v, want %v", got.Metadata(), tt.want.metadata)
				}
			} else {
				t.Errorf("ok() = %v, want %v", ok, true)
			}
		})
	}
}
