package client

import (
    "errors"
    "fmt"
    "time"
)

var ErrNotSupportedAuthorizationState = errors.New("not supported state")

type AuthorizationStateHandler interface {
    Handle(client *Client, state AuthorizationState) error
}

func Authorize(client *Client, authorizationStateHandler AuthorizationStateHandler) error {
    for {
        state, err := client.GetAuthorizationState()
        if err != nil {
            return err
        }

        err = authorizationStateHandler.Handle(client, state)
        if err != nil {
            return err
        }

        if state.AuthorizationStateType() == TypeAuthorizationStateReady {
            // dirty hack for db flush after authorization
            time.Sleep(1 * time.Second)
            return nil
        }
    }
}

type clientAuthorizer struct {
    TdlibParameters chan *TdlibParameters
    PhoneNumber     chan string
    Code            chan string
    State           chan AuthorizationState
    FirstName       chan string
    LastName        chan string
}

func ClientAuthorizer() *clientAuthorizer {
    return &clientAuthorizer{
        TdlibParameters: make(chan *TdlibParameters, 1),
        PhoneNumber:     make(chan string, 1),
        Code:            make(chan string, 1),
        State:           make(chan AuthorizationState, 10),
        FirstName:       make(chan string, 1),
        LastName:        make(chan string, 1),
    }
}

func (stateHandler *clientAuthorizer) Handle(client *Client, state AuthorizationState) error {
    stateHandler.State <- state

    switch state.AuthorizationStateType() {
    case TypeAuthorizationStateWaitTdlibParameters:
        _, err := client.SetTdlibParameters(<-stateHandler.TdlibParameters)
        return err

    case TypeAuthorizationStateWaitEncryptionKey:
        _, err := client.CheckDatabaseEncryptionKey(nil)
        return err

    case TypeAuthorizationStateWaitPhoneNumber:
        _, err := client.SetAuthenticationPhoneNumber(<-stateHandler.PhoneNumber, false, false)
        return err

    case TypeAuthorizationStateWaitCode:
        _, err := client.CheckAuthenticationCode(<-stateHandler.Code, <-stateHandler.FirstName, <-stateHandler.LastName)
        return err

    case TypeAuthorizationStateWaitPassword:
        return ErrNotSupportedAuthorizationState

    case TypeAuthorizationStateReady:
        close(stateHandler.TdlibParameters)
        close(stateHandler.PhoneNumber)
        close(stateHandler.Code)
        close(stateHandler.State)
        close(stateHandler.FirstName)
        close(stateHandler.LastName)

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

func CliInteractor(clientAuthorizer *clientAuthorizer, registration bool) {
    for {
        select {
        case state := <-clientAuthorizer.State:
            switch state.AuthorizationStateType() {
            case TypeAuthorizationStateWaitPhoneNumber:
                fmt.Println("Enter phone number: ")
                var phoneNumber string
                fmt.Scanln(&phoneNumber)

                clientAuthorizer.PhoneNumber <- phoneNumber

            case TypeAuthorizationStateWaitCode:
                var code string
                var firstName string
                var lastName string

                fmt.Println("Enter code: ")
                fmt.Scanln(&code)

                if registration {
                    fmt.Println("Enter first name: ")
                    fmt.Scanln(&firstName)

                    fmt.Println("Enter last name: ")
                    fmt.Scanln(&lastName)
                }

                clientAuthorizer.Code <- code
                clientAuthorizer.FirstName <- firstName
                clientAuthorizer.LastName <- lastName

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
        _, err := client.SetTdlibParameters(<-stateHandler.TdlibParameters)
        return err

    case TypeAuthorizationStateWaitEncryptionKey:
        _, err := client.CheckDatabaseEncryptionKey(nil)
        return err

    case TypeAuthorizationStateWaitPhoneNumber:
        _, err := client.CheckAuthenticationBotToken(<-stateHandler.Token)
        return err

    case TypeAuthorizationStateWaitCode:
        return ErrNotSupportedAuthorizationState

    case TypeAuthorizationStateWaitPassword:
        return ErrNotSupportedAuthorizationState

    case TypeAuthorizationStateReady:
        close(stateHandler.TdlibParameters)
        close(stateHandler.Token)
        close(stateHandler.State)

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
