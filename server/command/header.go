package command

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var (
	CommandHeader = model.Command{
		Trigger:          "header",
		AutoComplete:     true,
		AutoCompleteDesc: "Set the header of this channel.",
		AutoCompleteHint: "[New Header]",
		AutocompleteData: createCommandHeaderAutocompleteData(),
		DisplayName:      "Set Header",
	}
)

func createCommandHeaderAutocompleteData() *model.AutocompleteData {
	invite := model.NewAutocompleteData("header", "[New Header]", "Set the header of this channel.")
	invite.AddTextArgument("New Header", "[New Header]", "")
	return invite
}

func (r *Runner) Header() *model.CommandResponse {
	channel, err := r.pluginAPI.Channel.Get(r.args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			Text:         "Error updating channel header. Unable to retrieve channel.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	switch channel.Type {
	case model.CHANNEL_OPEN:
		if !r.pluginAPI.User.HasPermissionToChannel(r.args.UserId, r.args.ChannelId, model.PERMISSION_MANAGE_PUBLIC_CHANNEL_PROPERTIES) {
			return &model.CommandResponse{
				Text:         "You do not have permissions to modify channel header.",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}
	case model.CHANNEL_PRIVATE:
		if !r.pluginAPI.User.HasPermissionToChannel(r.args.UserId, r.args.ChannelId, model.PERMISSION_MANAGE_PRIVATE_CHANNEL_PROPERTIES) {
			return &model.CommandResponse{
				Text:         "You do not have permissions to modify channel header.",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}
	case model.CHANNEL_GROUP, model.CHANNEL_DIRECT:
		// Modifying the header is not linked to any specific permission for group/dm channels, so just check for membership.
		var channelMember *model.ChannelMember
		channelMember, err = r.pluginAPI.Channel.GetMember(r.args.ChannelId, r.args.UserId)
		if err != nil || channelMember == nil {
			return &model.CommandResponse{
				Text:         "You do not have permissions to modify channel header.",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}
	default:
		return &model.CommandResponse{
			Text:         "Unable to verify you have permission to modify channel header.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	oldHeader := channel.Header
	channel.Header = ""
	split := strings.SplitAfterN(r.args.Command, " ", 2)
	if len(split) == 2 {
		channel.Header = split[1]
	}

	user, err := r.pluginAPI.User.Get(r.args.UserId)
	if err != nil {
		return &model.CommandResponse{
			Text:         "Error updating channel header.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	message := ""
	if oldHeader == "" {
		message = fmt.Sprintf(r.args.T("api.channel.post_update_channel_header_message_and_forget.updated_to"), user.Username, channel.Header)
	} else if channel.Header == "" {
		message = fmt.Sprintf(r.args.T("api.channel.post_update_channel_header_message_and_forget.removed"), user.Username, oldHeader)
	} else {
		message = fmt.Sprintf(r.args.T("api.channel.post_update_channel_header_message_and_forget.updated_from"), user.Username, oldHeader, channel.Header)
	}

	if err := r.pluginAPI.Channel.Update(channel); err != nil {
		return &model.CommandResponse{
			Text:         "Error updating channel header.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	r.pluginAPI.Post.CreatePost(&model.Post{
		ChannelId: channel.Id,
		Message:   message,
		Type:      model.POST_HEADER_CHANGE,
		UserId:    r.args.UserId,
		Props: model.StringInterface{
			"username":   user.Username,
			"old_header": oldHeader,
			"new_header": channel.Header,
		},
	})

	return &model.CommandResponse{}
}
