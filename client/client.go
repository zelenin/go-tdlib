package client

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Client struct {
	jsonClient     *JsonClient
	extraGenerator ExtraGenerator
	responses      chan *Response
	listenerStore  *listenerStore
	catchersStore  *sync.Map
	updatesTimeout time.Duration
}

type Option func(*Client)

func WithExtraGenerator(extraGenerator ExtraGenerator) Option {
	return func(client *Client) {
		client.extraGenerator = extraGenerator
	}
}

func WithProxy(ctx context.Context, req *AddProxyRequest) Option {
	return func(client *Client) {
		_, _ = client.AddProxy(ctx, req)
	}
}

func WithLogVerbosity(req *SetLogVerbosityLevelRequest) Option {
	return func(client *Client) {
		_, _ = client.SetLogVerbosityLevel(req)
	}
}

func NewClient(options ...Option) *Client {
	client := &Client{
		jsonClient:    NewJsonClient(),
		responses:     make(chan *Response, 1000),
		listenerStore: newListenerStore(),
		catchersStore: &sync.Map{},
	}

	client.extraGenerator = UuidV4Generator()

	for _, option := range options {
		option(client)
	}

	tdlibInstance.addClient(client)

	go client.receiver()

	return client
}

func (client *Client) receiver() {
	for response := range client.responses {
		if response.Extra != "" {
			value, ok := client.catchersStore.Load(response.Extra)
			if ok {
				value.(chan *Response) <- response
			}
		}

		typ, err := UnmarshalType(response.Data)
		if err != nil {
			continue
		}

		needGc := false
		for _, listener := range client.listenerStore.Listeners() {
			if listener.IsActive() {
				listener.Updates <- typ
			} else {
				needGc = true
			}
		}
		if needGc {
			client.listenerStore.gc()
		}
	}
}

func (client *Client) Send(ctx context.Context, req Request) (*Response, error) {
	req.Extra = client.extraGenerator()

	catcher := make(chan *Response, 1)

	client.catchersStore.Store(req.Extra, catcher)

	defer func() {
		client.catchersStore.Delete(req.Extra)
		close(catcher)
	}()

	client.jsonClient.Send(req)

	select {
	case response := <-catcher:
		return response, nil

	case <-ctx.Done():
		return nil, errors.New("response catching timeout")
	}
}

func (client *Client) GetListener() *Listener {
	listener := &Listener{
		isActive: true,
		Updates:  make(chan Type, 1000),
	}
	client.listenerStore.Add(listener)

	return listener
}

func (client *Client) Stop(ctx context.Context) {
	_, _ = client.Destroy(ctx)
}
