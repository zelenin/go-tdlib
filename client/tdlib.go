package client

/*
#include <stdlib.h>
#include <td/telegram/td_json_client.h>

typedef void (*td_log_message_callback_ptr)(int, const char*);

extern void goLogMessageCallback(int verbosityLevel,  char* message);

static inline void setLogMessageCallback(int maxVerbosityLevel, td_log_message_callback_ptr callback) {
    td_set_log_message_callback(maxVerbosityLevel, callback);
}
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"unsafe"
)

var tdlibInstance *tdlib

func init() {
	tdlibInstance = &tdlib{
		timeout: 30,
		clients: map[int]*Client{},
	}
}

type tdlib struct {
	once    sync.Once
	timeout float64 // seconds
	mu      sync.Mutex
	clients map[int]*Client
}

func (instance *tdlib) addClient(client *Client) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	instance.clients[client.jsonClient.id] = client

	instance.once.Do(func() {
		go instance.receiver()
	})
}

func (instance *tdlib) getClient(id int) (*Client, error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	client, ok := instance.clients[id]
	if !ok {
		return nil, fmt.Errorf("client [id: %d] does not exist", id)
	}

	return client, nil
}

func (instance *tdlib) receiver() {
	for {
		resp, err := instance.receive(instance.timeout)
		if err != nil {
			continue
		}

		client, err := instance.getClient(resp.MetaClientId)
		if err != nil {
			log.Print(err)
			continue
		}

		client.responses <- resp

		typ, err := UnmarshalType(resp.Data)
		if err != nil {
			continue
		}
		if typ.GetConstructor() == ConstructorUpdateAuthorizationState &&
			typ.(*UpdateAuthorizationState).AuthorizationState.AuthorizationStateConstructor() == ConstructorAuthorizationStateClosed {
			close(client.responses)
			break
		}
	}
}

// Receives incoming updates and request responses from the TDLib client. May be called from any thread, but
// shouldn't be called simultaneously from two different threads.
// Returned pointer will be deallocated by TDLib during next call to td_json_client_receive or td_json_client_execute
// in the same thread, so it can't be used after that.
func (instance *tdlib) receive(timeout float64) (*Response, error) {
	result := C.td_receive(C.double(timeout))
	if result == nil {
		return nil, errors.New("update receiving timeout")
	}

	data := []byte(C.GoString(result))

	var resp Response
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	resp.Data = data

	return &resp, nil
}

func Execute(req Request) (*Response, error) {
	req.SetType(req.GetFunctionName())

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	query := C.CString(string(data))
	defer C.free(unsafe.Pointer(query))
	result := C.td_execute(query)
	if result == nil {
		return nil, errors.New("request can't be parsed")
	}

	data = []byte(C.GoString(result))

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	resp.Data = data

	return &resp, nil
}

type JsonClient struct {
	id int
}

func NewJsonClient() *JsonClient {
	return &JsonClient{
		id: int(C.td_create_client_id()),
	}
}

// Sends request to the TDLib client. May be called from any thread.
func (jsonClient *JsonClient) Send(req Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	query := C.CString(string(data))
	defer C.free(unsafe.Pointer(query))

	C.td_send(C.int(jsonClient.id), query)

	return nil
}

// Synchronously executes TDLib request. May be called from any thread.
// Only a few requests can be executed synchronously.
// Returned pointer will be deallocated by TDLib during next call to td_json_client_receive or td_json_client_execute
// in the same thread, so it can't be used after that.
func (jsonClient *JsonClient) Execute(req Request) (*Response, error) {
	return Execute(req)
}

type meta struct {
	MetaType     string `json:"@type"`
	MetaExtra    string `json:"@extra"`
	MetaClientId int    `json:"@client_id"`
}

type reqMeta struct {
	MetaType  string `json:"@type"`
	MetaExtra string `json:"@extra"`
}

type Request interface {
	GetFunctionName() string
	SetExtra(extra string)
	GetExtra() string
	SetType(typ string)
	GetType() string
}

type request struct {
	reqMeta
}

func (req *request) SetExtra(extra string) {
	req.MetaExtra = extra
}

func (req *request) GetExtra() string {
	return req.MetaExtra
}

func (req *request) SetType(typ string) {
	req.MetaType = typ
}

func (req *request) GetType() string {
	return req.MetaType
}

type Response struct {
	meta
	Data json.RawMessage
}

type ResponseError struct {
	Err *Error
}

func (responseError ResponseError) Error() string {
	return fmt.Sprintf("%d %s", responseError.Err.Code, responseError.Err.Message)
}

func buildResponseError(data json.RawMessage) error {
	respErr, err := UnmarshalError(data)
	if err != nil {
		return err
	}

	return ResponseError{
		Err: respErr,
	}
}

// JsonInt64 alias for int64, in order to deal with json big number problem
type JsonInt64 int64

// MarshalJSON marshals to json
func (jsonInt64 JsonInt64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(jsonInt64), 10) + `"`), nil
}

// UnmarshalJSON unmarshals from json
func (jsonInt64 *JsonInt64) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	jsonBigInt, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}

	*jsonInt64 = JsonInt64(jsonBigInt)

	return nil
}

type Type interface {
	GetConstructor() string
	GetType() string
}

var (
	logCallback func(int, string)
)

//export goLogMessageCallback
func goLogMessageCallback(verbosityLevel C.int, message *C.char) {
	if logCallback != nil {
		logCallback(int(verbosityLevel), C.GoString(message))
	}
}

// Sets the callback that will be called when a message is added to the internal TDLib log.
// None of the TDLib methods can be called from the callback.
// By default the callback is not set.
func SetLogMessageCallback(maxVerbosityLevel int, callback func(verbosityLevel int, message string)) {
	if callback == nil {
		logCallback = nil
		C.setLogMessageCallback(C.int(maxVerbosityLevel), nil)
	} else {
		logCallback = callback
		C.setLogMessageCallback(C.int(maxVerbosityLevel), (C.td_log_message_callback_ptr)(C.goLogMessageCallback))
	}
}
