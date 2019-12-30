package client

import (
	"errors"
	"sync"
	"time"
)

type Client struct {
	jsonClient     *JsonClient
	extraGenerator ExtraGenerator
	catcher        chan *Response
	listenerStore  *listenerStore
	catchersStore  *sync.Map
	updatesTimeout time.Duration
	catchTimeout   time.Duration
}

type Option func(*Client)

func WithExtraGenerator(extraGenerator ExtraGenerator) Option {
	return func(client *Client) {
		client.extraGenerator = extraGenerator
	}
}

func WithCatchTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.catchTimeout = timeout
	}
}

func WithUpdatesTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.updatesTimeout = timeout
	}
}

func WithProxy(req *AddProxyRequest) Option {
	return func(client *Client) {
		client.AddProxy(req)
	}
}

func WithLogVerbosity(req *SetLogVerbosityLevelRequest) Option {
	return func(client *Client) {
		client.SetLogVerbosityLevel(req)
	}
}

func NewClient(authorizationStateHandler AuthorizationStateHandler, options ...Option) (*Client, error) {
	catchersListener := make(chan *Response, 1000)

	client := &Client{
		jsonClient:    NewJsonClient(),
		catcher:       catchersListener,
		listenerStore: newListenerStore(),
		catchersStore: &sync.Map{},
	}

	client.extraGenerator = UuidV4Generator()
	client.catchTimeout = 60 * time.Second
	client.updatesTimeout = 60 * time.Second

	for _, option := range options {
		option(client)
	}

	go client.receive()
	go client.catch(catchersListener)

	err := Authorize(client, authorizationStateHandler)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) receive() {
	for {
		resp, err := client.jsonClient.Receive(client.updatesTimeout)
		if err != nil {
			continue
		}
		client.catcher <- resp

		typ, err := UnmarshalType(resp.Data)
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

func (client *Client) catch(updates chan *Response) {
	for update := range updates {
		if update.Extra != "" {
			value, ok := client.catchersStore.Load(update.Extra)
			if ok {
				value.(chan *Response) <- update
			}
		}
	}
}

func (client *Client) Send(req Request) (*Response, error) {
	req.Extra = client.extraGenerator()

	catcher := make(chan *Response, 1)

	client.catchersStore.Store(req.Extra, catcher)

	defer func() {
		close(catcher)
		client.catchersStore.Delete(req.Extra)
	}()

	client.jsonClient.Send(req)

	select {
	case response := <-catcher:
		return response, nil

	case <-time.After(client.catchTimeout):
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

func (client *Client) Stop() {
	client.Destroy()
	client.jsonClient.Destroy()
}
