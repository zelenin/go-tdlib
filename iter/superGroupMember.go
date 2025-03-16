package iter

import (
	"context"
	"github.com/zelenin/go-tdlib/client"
	"iter"
)

func SuperGroupMemberIter(ctx context.Context, tdlibClient *client.Client, supergroupId int64) iter.Seq2[*client.ChatMember, error] {
	return func(yield func(*client.ChatMember, error) bool) {
		const limit = 200
		var offset int32 = 0

		for {
			chatMembers, err := tdlibClient.GetSupergroupMembers(context.Background(), &client.GetSupergroupMembersRequest{
				SupergroupId: supergroupId,
				Filter:       nil,
				Offset:       offset,
				Limit:        limit,
			})
			if err != nil {
				yield(nil, err)
				return
			}

			for _, chatMember := range chatMembers.Members {
				if !yield(chatMember, nil) {
					return
				}
			}

			if chatMembers.TotalCount == 0 || len(chatMembers.Members) == 0 {
				return
			}

			offset += int32(len(chatMembers.Members))
		}
	}
}
