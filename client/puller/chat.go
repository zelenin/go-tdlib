package puller

import (
	"github.com/zelenin/go-tdlib/client"
)

func ChatHistory(tdlibClient *client.Client, chatId int64) (chan *client.Message, chan error) {
	messageChan := make(chan *client.Message, 10)
	errChan := make(chan error, 1)

	var fromMessageId int64 = 0
	var offset int32 = 0
	var limit int32 = 100

	go chatHistory(tdlibClient, messageChan, errChan, chatId, fromMessageId, offset, limit, false)

	return messageChan, errChan
}

func chatHistory(tdlibClient *client.Client, messageChan chan *client.Message, errChan chan error, chatId int64, fromMessageId int64, offset int32, limit int32, onlyLocal bool) {
	defer func() {
		close(messageChan)
		close(errChan)
	}()

	for {
		messages, err := tdlibClient.GetChatHistory(&client.GetChatHistoryRequest{
			ChatId:        chatId,
			FromMessageId: fromMessageId,
			Offset:        offset,
			Limit:         limit,
			OnlyLocal:     onlyLocal,
		})
		if err != nil {
			errChan <- err

			return
		}

		if len(messages.Messages) == 0 {
			errChan <- EOP

			break
		}

		for _, message := range messages.Messages {
			fromMessageId = message.Id

			messageChan <- message
		}
	}
}
