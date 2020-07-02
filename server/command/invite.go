package command

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var (
	CommandInvite = model.Command{
		Trigger:          "invite",
		AutoComplete:     true,
		AutoCompleteDesc: "Invite users to this channel.",
		AutoCompleteHint: "@[username]",
		AutocompleteData: createInviteAutocompleteData(),
		DisplayName:      "Invite",
	}
)

func createInviteAutocompleteData() *model.AutocompleteData {
	invite := model.NewAutocompleteData("invite", "@[users]", "Invite users to the current channel.")
	invite.AddTextArgument("Users to invite", "@[users]", "")
	return invite
}

func (r *Runner) Invite() *model.CommandResponse {
	split := strings.Fields(r.args.Command)
	if len(split) < 2 {
		return &model.CommandResponse{
			Text:         "Users not specified.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	channel, err := r.pluginAPI.Channel.Get(r.args.ChannelId)
	if err != nil {
		return &model.CommandResponse{
			Text:         "Failed to get channel.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	// Permissions Check
	switch channel.Type {
	case model.CHANNEL_OPEN:
		if !r.pluginAPI.User.HasPermissionToChannel(r.args.UserId, channel.Id, model.PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS) {
			return &model.CommandResponse{
				Text:         "You don't have permissions to invite users to this chanel",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}
	case model.CHANNEL_PRIVATE:
		if !r.pluginAPI.User.HasPermissionToChannel(r.args.UserId, channel.Id, model.PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS) {
			return &model.CommandResponse{
				Text:         "You don't have permissions to invite users to this chanel",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}
	default:
		return &model.CommandResponse{
			Text:         "You can't invite additional users to direct channels.",
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		}
	}

	users := split[1:]
	for _, user := range users {
		username := strings.TrimPrefix(user, "@")

		user, err := r.pluginAPI.User.GetByUsername(username)
		if err != nil {
			return &model.CommandResponse{
				Text:         "Failed to get by username.",
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			}
		}

		r.pluginAPI.Channel.AddUser(r.args.ChannelId, user.Id, r.args.UserId)

	}

	return &model.CommandResponse{}
}
