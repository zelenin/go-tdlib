package client

import "strings"

func applyFilters(message *Message, filters []string) bool {
	c := 0
	for _, f := range filters {
		if filter[f](message) {
			c++
		}
	}
	if c == len(filters) {
		return true
	}
	return false
}

func isValidCommand(message *Message, cmd, prefix string) []string {
	rawText := strings.Trim(message.Content.(*MessageText).Text.Text, " ")
	args := strings.Fields(rawText)
	command := strings.Trim(prefix+cmd, " ")
	if args[0] == command {
		return args
	}
	return nil
}

func notMe(message *Message) bool {
	if message.SenderUserId != clientID {
		return true
	}
	return false
}

func me(message *Message) bool {
	if message.SenderUserId == clientID {
		return true
	}
	return false
}

func incoming(message *Message) bool {
	if !message.IsOutgoing {
		return true
	}
	return false
}

func outgoing(message *Message) bool {
	if message.IsOutgoing {
		return true
	}
	return false
}

func text(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageText {
		return true
	}
	return false
}

func reply(message *Message) bool {
	if message.ReplyToMessageId != 0 {
		return true
	}
	return false
}

func forwarded(message *Message) bool {
	if message.ForwardInfo != nil {
		return true
	}
	return false
}

func caption(message *Message) bool {
	var caption string
	switch message.Content.MessageContentType() {
	case TypeMessageAudio:
		caption = message.Content.(*MessageAudio).Caption.Text
	case TypeMessageVideo:
		caption = message.Content.(*MessageVideo).Caption.Text
	case TypeMessageAnimation:
		caption = message.Content.(*MessageAnimation).Caption.Text
	case TypeMessageDocument:
		caption = message.Content.(*MessageDocument).Caption.Text
	case TypeMessagePhoto:
		caption = message.Content.(*MessagePhoto).Caption.Text
	case TypeMessageVoiceNote:
		caption = message.Content.(*MessageVoiceNote).Caption.Text
	}

	if caption != "" {
		return true
	}
	return false

}

func edited(message *Message) bool {
	if message.EditDate != 0 {
		return true
	}
	return false
}

func audio(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageAudio {
		return true
	}
	return false
}

func document(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageDocument {
		return true
	}
	return false
}

func photo(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessagePhoto {
		return true
	}
	return false
}

func sticker(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageSticker {
		return true
	}
	return false

}

func animation(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageAnimation {
		return true
	}
	return false
}

func game(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageGame {
		return true
	}
	return false
}

func video(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageVideo {
		return true
	}
	return false
}

func voice(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageVoiceNote {
		return true
	}
	return false
}

func videoNote(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageVideoNote {
		return true
	}
	return false
}

func contact(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageContact {
		return true
	}
	return false
}

func location(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageLocation {
		return true
	}
	return false
}

func venue(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessageVenue {
		return true
	}
	return false
}

func poll(message *Message) bool {
	if message.Content.MessageContentType() == TypeMessagePoll {
		return true
	}
	return false
}

func channel(message *Message) bool {
	return message.IsChannelPost
}

func media(message *Message) bool {
	switch message.Content.MessageContentType() {
	case TypeMessageAudio:
		return true
	case TypeMessageVideo:
		return true
	case TypeMessageAnimation:
		return true
	case TypeMessageDocument:
		return true
	case TypeMessagePhoto:
		return true
	case TypeMessageVoiceNote:
		return true
	default:
		return false
	}

}

var filter = map[string]func(*Message) bool{
	"FilterNotMe":     notMe,
	"FilterMe":        me,
	"FilterIncoming":  incoming,
	"FilterOutgoing":  outgoing,
	"FilterText":      text,
	"FilterReply":     reply,
	"FilterForwarded": forwarded,
	"FilterCaption":   caption,
	"FilterEdited":    edited,
	"FilterAudo":      audio,
	"FilterDocument":  document,
	"FilterPhoto":     photo,
	"FilterSticker":   sticker,
	"FilterAnimation": animation,
	"FilterGame":      game,
	"FilterVideo":     video,
	"FilterVoice":     voice,
	"FilterVideoNote": videoNote,
	"FilterContact":   contact,
	"FilterLocation":  location,
	"FilterVenue":     venue,
	"FilterPoll":      poll,
	"FilterChannel":   channel,
	"FilterMedia":     media,
}

// Filters
const (
	FilterNotMe      = "FilterNotMe"      // messages that aren't generated by you yourself.
	FilterMe         = "FilterMe"         // messages generated by you yourself.
	FilterIncoming   = "FilterIncoming"   // incoming messages. Messages sent to your own chat (Saved Messages) are also recognised as incoming.
	FilterOutgoing   = "FilterOutgoing"   // outgoing messages. Messages sent to your own chat (Saved Messages) are not recognized as outgoing.
	FilterText       = "FilterText"       // Filter text messages.
	FilterReply      = "FilterReply"      // messages that are replies to other messages.
	FilterForwarded  = "FilterForwarded"  // messages that are forwarded.
	FilterCaption    = "FilterCaption"    // media messages that contain captions.
	FilterEdited     = "FilterEdited"     // Filter edited messages.
	FilterAudo       = "FilterAudo"       // messages that contain Audio objects.
	FilterDocument   = "FilterDocument"   // messages that contain Document objects.
	FilterPhoto      = "FilterPhoto"      // messages that contain Photo objects.
	FilterSticker    = "FilterSticker"    // messages that contain Sticker objects.
	FilterAnimation  = "FilterAnimation"  // messages that contain Animation objects.
	FilterGame       = "FilterGame"       // messages that contain Game objects.
	FilterVideo      = "FilterVideo"      // messages that contain Video objects.
	FilterMediaGroup = "FilterMediaGroup" // messages containing photos or videos being part of an album.
	FilterVoice      = "FilterVoice"      // messages that contain Voice note objects.
	FilterVideoNote  = "FilterVideoNote"  // messages that contain VideoNote objects.
	FilterContact    = "FilterContact"    // messages that contain Contact objects.
	FilterLocation   = "FilterLocation"   // messages that contain Location objects.
	FilterVenue      = "FilterVenue"      // messages that contain Venue objects.
	FilterWebPage    = "FilterWebPage"    // messages sent with a webpage preview.
	FilterPool       = "FilterPool"       // messages that contain Poll objects.
	FilterPrivate    = "FilterPrivate"    // messages sent in private chats.
	FilterGroup      = "FilterGroup"      // messages sent in group or supergroup chats.
	FilterChannel    = "FilterChannel"    // messages sent in channels.
	FilterMedia      = "FilterMedia"      // media messages.
)
