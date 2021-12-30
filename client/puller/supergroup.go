package puller

import (
	"github.com/zelenin/go-tdlib/client"
)

func SupergroupMembers(tdlibClient *client.Client, supergroupId int64) (chan *client.ChatMember, chan error) {
	chatMemberChan := make(chan *client.ChatMember, 10)
	errChan := make(chan error, 1)

	var filter client.SupergroupMembersFilter = nil
	var offset int32 = 0
	var limit int32 = 200

	go supergroupMembers(tdlibClient, chatMemberChan, errChan, supergroupId, filter, offset, limit)

	return chatMemberChan, errChan
}

func supergroupMembers(tdlibClient *client.Client, chatMemberChan chan *client.ChatMember, errChan chan error, supergroupId int64, filter client.SupergroupMembersFilter, offset int32, limit int32) {
	defer func() {
		close(chatMemberChan)
		close(errChan)
	}()

	var page int32 = 0

	for {
		chatMembers, err := tdlibClient.GetSupergroupMembers(&client.GetSupergroupMembersRequest{
			SupergroupId: supergroupId,
			Filter:       filter,
			Offset:       page*limit + offset,
			Limit:        limit,
		})
		if err != nil {
			errChan <- err

			return
		}

		if len(chatMembers.Members) == 0 {
			errChan <- EOP

			break
		}

		for _, member := range chatMembers.Members {
			chatMemberChan <- member
		}

		page++
	}
}
