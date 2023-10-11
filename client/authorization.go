package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotSupportedAuthorizationState = errors.New("not supported state")

type AuthorizationStateHandler interface {
	Handle(context.Context, *Client, AuthorizationState) error
	Close()
}

func Authorize(
	ctx context.Context,
	client *Client,
	handler AuthorizationStateHandler,
) error {
	defer handler.Close()

	var authorizationError error

	for {
		state, err := client.GetAuthorizationState(ctx)
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

		err = handler.Handle(ctx, client, state)
		if err != nil {
			authorizationError = err
			_, _ = client.Close(ctx)
		}
	}
}

type AppAuthorizer struct {
	TdlibParameters chan *SetTdlibParametersRequest
	PhoneNumber     chan string
	Code            chan string
	State           chan AuthorizationState
	Password        chan string
}

func NewAppAuthorizer() *AppAuthorizer {
	return &AppAuthorizer{
		TdlibParameters: make(chan *SetTdlibParametersRequest, 1),
		PhoneNumber:     make(chan string, 1),
		Code:            make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
		Password:        make(chan string, 1),
	}
}

func (stateHandler *AppAuthorizer) Handle(
	ctx context.Context,
	client *Client,
	state AuthorizationState,
) error {
	stateHandler.State <- state

	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(ctx, <-stateHandler.TdlibParameters)
		return err

	case TypeAuthorizationStateWaitEmailAddress:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitEmailCode:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitPhoneNumber:
		req := &SetAuthenticationPhoneNumberRequest{
			PhoneNumber: <-stateHandler.PhoneNumber,
			Settings: &PhoneNumberAuthenticationSettings{
				AllowFlashCall:       false,
				IsCurrentPhoneNumber: false,
				AllowSmsRetrieverApi: false,
			},
		}
		_, err := client.SetAuthenticationPhoneNumber(ctx, req)
		return err

	case TypeAuthorizationStateWaitCode:
		req := &CheckAuthenticationCodeRequest{
			Code: <-stateHandler.Code,
		}
		_, err := client.CheckAuthenticationCode(ctx, req)
		return err

	case TypeAuthorizationStateWaitRegistration:
		return ErrNotSupportedAuthorizationState

	case TypeAuthorizationStateWaitPassword:
		req := &CheckAuthenticationPasswordRequest{
			Password: <-stateHandler.Password,
		}
		_, err := client.CheckAuthenticationPassword(ctx, req)
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

func (stateHandler *AppAuthorizer) Close() {
	close(stateHandler.TdlibParameters)
	close(stateHandler.PhoneNumber)
	close(stateHandler.Code)
	close(stateHandler.State)
	close(stateHandler.Password)
}

func CliInteractor(clientAuthorizer *AppAuthorizer) {
	for {
		select {
		case state, ok := <-clientAuthorizer.State:
			if !ok {
				return
			}

			switch state.AuthorizationStateType() {
			case TypeAuthorizationStateWaitPhoneNumber:
				fmt.Println("Enter phone number:")
				var phoneNumber string
				_, _ = fmt.Scanln(&phoneNumber)

				clientAuthorizer.PhoneNumber <- phoneNumber

			case TypeAuthorizationStateWaitCode:
				var code string

				fmt.Println("Enter code:")
				_, _ = fmt.Scanln(&code)

				clientAuthorizer.Code <- code

			case TypeAuthorizationStateWaitPassword:
				fmt.Println("Enter password:")
				var password string
				_, _ = fmt.Scanln(&password)

				clientAuthorizer.Password <- password

			case TypeAuthorizationStateReady:
				return
			}
		}
	}
}

type BotAuthorizer struct {
	TdlibParameters chan *SetTdlibParametersRequest
	Token           chan string
	State           chan AuthorizationState
}

func NewBotAuthorizer(token string) *BotAuthorizer {
	botAuthorizer := &BotAuthorizer{
		TdlibParameters: make(chan *SetTdlibParametersRequest, 1),
		Token:           make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
	}

	botAuthorizer.Token <- token

	return botAuthorizer
}

func (stateHandler *BotAuthorizer) Handle(
	ctx context.Context,
	client *Client,
	state AuthorizationState,
) error {
	stateHandler.State <- state

	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(ctx, <-stateHandler.TdlibParameters)
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		req := &CheckAuthenticationBotTokenRequest{
			Token: <-stateHandler.Token,
		}
		_, err := client.CheckAuthenticationBotToken(ctx, req)
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

func (stateHandler *BotAuthorizer) Close() {
	close(stateHandler.TdlibParameters)
	close(stateHandler.Token)
	close(stateHandler.State)
}
