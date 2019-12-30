package client

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotSupportedAuthorizationState = errors.New("not supported state")

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
	TdlibParameters chan *TdlibParameters
	PhoneNumber     chan string
	Code            chan string
	State           chan AuthorizationState
	Password        chan string
}

func ClientAuthorizer() *clientAuthorizer {
	return &clientAuthorizer{
		TdlibParameters: make(chan *TdlibParameters, 1),
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
		_, err := client.SetTdlibParameters(&SetTdlibParametersRequest{
			Parameters: <-stateHandler.TdlibParameters,
		})
		return err

	case TypeAuthorizationStateWaitEncryptionKey:
		_, err := client.CheckDatabaseEncryptionKey(&CheckDatabaseEncryptionKeyRequest{})
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

	case TypeAuthorizationStateWaitCode:
		_, err := client.CheckAuthenticationCode(&CheckAuthenticationCodeRequest{
			Code: <-stateHandler.Code,
		})
		return err

	case TypeAuthorizationStateWaitRegistration:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitPassword:
		_, err := client.CheckAuthenticationPassword(&CheckAuthenticationPasswordRequest{
			Password: <-stateHandler.Password,
		})
		return err

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateClosing:
		return nil

	case TypeAuthorizationStateClosed:
		return nil
	}

	return ErrNotSupportedAuthorizationState
}

func (stateHandler *clientAuthorizer) Close() {
	close(stateHandler.TdlibParameters)
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
	TdlibParameters chan *TdlibParameters
	Token           chan string
	State           chan AuthorizationState
}

func BotAuthorizer(token string) *botAuthorizer {
	botAuthorizer := &botAuthorizer{
		TdlibParameters: make(chan *TdlibParameters, 1),
		Token:           make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
	}

	botAuthorizer.Token <- token

	return botAuthorizer
}

func (stateHandler *botAuthorizer) Handle(client *Client, state AuthorizationState) error {
	stateHandler.State <- state

	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(&SetTdlibParametersRequest{
			Parameters: <-stateHandler.TdlibParameters,
		})
		return err

	case TypeAuthorizationStateWaitEncryptionKey:
		_, err := client.CheckDatabaseEncryptionKey(&CheckDatabaseEncryptionKeyRequest{})
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.CheckAuthenticationBotToken(&CheckAuthenticationBotTokenRequest{
			Token: <-stateHandler.Token,
		})
		return err

	case TypeAuthorizationStateWaitCode:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitPassword:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateClosing:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateClosed:
		return ErrNotSupportedAuthorizationState
	}

	return ErrNotSupportedAuthorizationState
}

func (stateHandler *botAuthorizer) Close() {
	close(stateHandler.TdlibParameters)
	close(stateHandler.Token)
	close(stateHandler.State)
}
