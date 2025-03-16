package iter

import (
	"context"
	"github.com/zelenin/go-tdlib/client"
	"iter"
)

func ChatHistory(ctx context.Context, tdlibClient *client.Client, chatId int64) iter.Seq2[*client.Message, error] {
	return func(yield func(*client.Message, error) bool) {
		const offset int32 = 0
		const limit int32 = 100
		const onlyLocal = false

		var fromMessageId int64 = 0

		for {
			messages, err := tdlibClient.GetChatHistory(ctx, &client.GetChatHistoryRequest{
				ChatId:        chatId,
				FromMessageId: fromMessageId,
				Offset:        offset,
				Limit:         limit,
				OnlyLocal:     onlyLocal,
			})
			if err != nil {
				yield(nil, err)
				return
			}

			for _, message := range messages.Messages {
				fromMessageId = message.Id
				if !yield(message, nil) {
					return
				}
			}

			if len(messages.Messages) == 0 {
				return
			}
		}
	}
}
