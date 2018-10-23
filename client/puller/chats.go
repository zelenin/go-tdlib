package puller

import (
	"math"

	"github.com/zelenin/go-tdlib/client"
)

func Chats(tdlibClient *client.Client) (chan *client.Chat, chan error) {
	chatChan := make(chan *client.Chat, 10)
	errChan := make(chan error, 1)

	var offsetOrder client.JsonInt64 = math.MaxInt64
	var offsetChatId int64 = 0
	var limit int32 = 100

	go chats(tdlibClient, chatChan, errChan, offsetOrder, offsetChatId, limit)

	return chatChan, errChan
}

func chats(tdlibClient *client.Client, chatChan chan *client.Chat, errChan chan error, offsetOrder client.JsonInt64, offsetChatId int64, limit int32) {
	defer func() {
		close(chatChan)
		close(errChan)
	}()

	for {
		chats, err := tdlibClient.GetChats(&client.GetChatsRequest{
			OffsetOrder:  offsetOrder,
			OffsetChatId: offsetChatId,
			Limit:        limit,
		})
		if err != nil {
			errChan <- err

			return
		}

		if len(chats.ChatIds) == 0 {
			errChan <- EOP

			break
		}

		for _, chatId := range chats.ChatIds {
			chat, err := tdlibClient.GetChat(&client.GetChatRequest{
				ChatId: chatId,
			})
			if err != nil {
				errChan <- err

				return
			}

			offsetOrder = chat.Order
			offsetChatId = chat.Id

			chatChan <- chat
		}
	}
}
