package client

import (
	"fmt"
	"time"
)

type notSupportedAuthorizationState struct {
	state AuthorizationState
}

func (err *notSupportedAuthorizationState) Error() string {
	return fmt.Sprintf("not supported authorization state: %s", err.state.AuthorizationStateType())
}

func NotSupportedAuthorizationState(state AuthorizationState) error {
	return &notSupportedAuthorizationState{
		state: state,
	}
}

type AuthorizationStateHandler interface {
	Handle(client *Client, state AuthorizationState) error
	Close()
}

func Authorize(client *Client, authorizationStateHandler AuthorizationStateHandler) error {
	defer authorizationStateHandler.Close()

	var authorizationError error

	for {
		state, err := client.GetAuthorizationState()
		if err != nil {
			return err
		}

		if state.AuthorizationStateType() == TypeAuthorizationStateClosed {
			return authorizationError
		}

		if state.AuthorizationStateType() == TypeAuthorizationStateReady {
			// dirty hack for db flush after authorization
			time.Sleep(1 * time.Second)
			return nil
		}

		err = authorizationStateHandler.Handle(client, state)
		if err != nil {
			authorizationError = err
			client.Close()
		}
	}
}

type clientAuthorizer struct {
	TdlibParameters *SetTdlibParametersRequest
	PhoneNumber     chan string
	Code            chan string
	State           chan AuthorizationState
	Password        chan string
}

func ClientAuthorizer(tdlibParameters *SetTdlibParametersRequest) *clientAuthorizer {
	return &clientAuthorizer{
		TdlibParameters: tdlibParameters,
		PhoneNumber:     make(chan string, 1),
		Code:            make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
		Password:        make(chan string, 1),
	}
}

func (stateHandler *clientAuthorizer) Handle(client *Client, state AuthorizationState) error {
	stateHandler.State <- state

	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(stateHandler.TdlibParameters)
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.SetAuthenticationPhoneNumber(&SetAuthenticationPhoneNumberRequest{
			PhoneNumber: <-stateHandler.PhoneNumber,
			Settings: &PhoneNumberAuthenticationSettings{
				AllowFlashCall:       false,
				IsCurrentPhoneNumber: false,
				AllowSmsRetrieverApi: false,
			},
		})
		return err

	case TypeAuthorizationStateWaitEmailAddress:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitEmailCode:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitCode:
		_, err := client.CheckAuthenticationCode(&CheckAuthenticationCodeRequest{
			Code: <-stateHandler.Code,
		})
		return err

	case TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitRegistration:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitPassword:
		_, err := client.CheckAuthenticationPassword(&CheckAuthenticationPasswordRequest{
			Password: <-stateHandler.Password,
		})
		return err

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateClosing:
		return nil

	case TypeAuthorizationStateClosed:
		return nil
	}

	return NotSupportedAuthorizationState(state)
}

func (stateHandler *clientAuthorizer) Close() {
	close(stateHandler.PhoneNumber)
	close(stateHandler.Code)
	close(stateHandler.State)
	close(stateHandler.Password)
}

func CliInteractor(clientAuthorizer *clientAuthorizer) {
	for {
		select {
		case state, ok := <-clientAuthorizer.State:
			if !ok {
				return
			}

			switch state.AuthorizationStateType() {
			case TypeAuthorizationStateWaitPhoneNumber:
				fmt.Println("Enter phone number: ")
				var phoneNumber string
				fmt.Scanln(&phoneNumber)

				clientAuthorizer.PhoneNumber <- phoneNumber

			case TypeAuthorizationStateWaitCode:
				var code string

				fmt.Println("Enter code: ")
				fmt.Scanln(&code)

				clientAuthorizer.Code <- code

			case TypeAuthorizationStateWaitPassword:
				fmt.Println("Enter password: ")
				var password string
				fmt.Scanln(&password)

				clientAuthorizer.Password <- password

			case TypeAuthorizationStateReady:
				return
			}
		}
	}
}

type botAuthorizer struct {
	tdlibParameters *SetTdlibParametersRequest
	token           string
}

func BotAuthorizer(tdlibParameters *SetTdlibParametersRequest, token string) *botAuthorizer {
	return &botAuthorizer{
		tdlibParameters: tdlibParameters,
		token:           token,
	}
}

func (stateHandler *botAuthorizer) Handle(client *Client, state AuthorizationState) error {
	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(stateHandler.tdlibParameters)
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.CheckAuthenticationBotToken(&CheckAuthenticationBotTokenRequest{
			Token: stateHandler.token,
		})
		return err

	case TypeAuthorizationStateWaitEmailAddress:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitEmailCode:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitCode:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitRegistration:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitPassword:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateClosing:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateClosed:
		return NotSupportedAuthorizationState(state)
	}

	return NotSupportedAuthorizationState(state)
}

func (stateHandler *botAuthorizer) Close() {}

type qrAuthorizer struct {
	TdlibParameters *SetTdlibParametersRequest
	Password        chan string
	lastLink        string
	LinkHandler     func(link string) error
}

func QrAuthorizer(tdlibParameters *SetTdlibParametersRequest, linkHandler func(link string) error) *qrAuthorizer {
	stateHandler := &qrAuthorizer{
		TdlibParameters: tdlibParameters,
		Password:        make(chan string, 1),
		LinkHandler:     linkHandler,
	}

	return stateHandler
}

func (stateHandler *qrAuthorizer) Handle(client *Client, state AuthorizationState) error {
	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(stateHandler.TdlibParameters)
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.RequestQrCodeAuthentication(&RequestQrCodeAuthenticationRequest{})
		return err

	case TypeAuthorizationStateWaitOtherDeviceConfirmation:
		link := state.(*AuthorizationStateWaitOtherDeviceConfirmation).Link

		if link == stateHandler.lastLink {
			return nil
		}

		err := stateHandler.LinkHandler(link)
		if err != nil {
			return err
		}

		stateHandler.lastLink = link

		return nil

	case TypeAuthorizationStateWaitCode:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateWaitPassword:
		_, err := client.CheckAuthenticationPassword(&CheckAuthenticationPasswordRequest{
			Password: <-stateHandler.Password,
		})
		return err

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateClosing:
		return NotSupportedAuthorizationState(state)

	case TypeAuthorizationStateClosed:
		return NotSupportedAuthorizationState(state)
	}

	return NotSupportedAuthorizationState(state)
}

func (stateHandler *qrAuthorizer) Close() {
	close(stateHandler.Password)
}
