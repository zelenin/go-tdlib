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
    listeners      []chan Type
    catchersStore  *sync.Map
}

type Option func(*Client)

func WithExtraGenerator(extraGenerator ExtraGenerator) Option {
    return func(client *Client) {
        client.extraGenerator = extraGenerator
    }
}

func WithListener(listener chan Type) Option {
    return func(client *Client) {
        client.listeners = append(client.listeners, listener)
    }
}

func NewClient(authorizationStateHandler AuthorizationStateHandler, options ...Option) (*Client, error) {
    catchersListener := make(chan *Response, 1000)

    client := &Client{
        jsonClient:    NewJsonClient(),
        catcher:       catchersListener,
        listeners:     []chan Type{},
        catchersStore: &sync.Map{},
    }

    for _, option := range options {
        option(client)
    }

    if client.extraGenerator == nil {
        client.extraGenerator = UuidV4Generator()
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
        resp, err := client.jsonClient.Receive(10)
        if err != nil {
            continue
        }
        client.catcher <- resp

        typ, err := UnmarshalType(resp.Data)
        if err != nil {
            continue
        }

        for _, listener := range client.listeners {
            listener <- typ
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

    case <-time.After(10 * time.Second):
        return nil, errors.New("timeout")
    }
}
