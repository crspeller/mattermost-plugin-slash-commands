package command

import (
	"errors"
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Register is a function that allows the runner to register commands with the mattermost server.
type Register func(*model.Command) error

// RegisterCommands should be called by the plugin to register all necessary commands
func RegisterCommands(registerFunc Register) error {
	registerFunc(&CommandInvite)
	registerFunc(&CommandHeader)

	return nil
}

// Runner handles commands.
type Runner struct {
	context   *plugin.Context
	args      *model.CommandArgs
	pluginAPI *pluginapi.Client
}

// NewCommandRunner creates a command runner.
func NewCommandRunner(ctx *plugin.Context, args *model.CommandArgs, api *pluginapi.Client) *Runner {
	return &Runner{
		context:   ctx,
		args:      args,
		pluginAPI: api,
	}
}

func (r *Runner) isValid() error {
	if r.context == nil || r.args == nil || r.pluginAPI == nil {
		return errors.New("invalid arguments to command.Runner")
	}
	return nil
}

func (r *Runner) Execute() (*model.CommandResponse, error) {
	if err := r.isValid(); err != nil {
		return nil, err
	}

	split := strings.Fields(r.args.Command)
	cmd := split[0]
	cmd = strings.TrimPrefix(cmd, "/")

	var resp *model.CommandResponse
	switch cmd {
	case CommandInvite.Trigger:
		resp = r.Invite()
	case CommandHeader.Trigger:
		resp = r.Header()
	}

	return resp, nil
}
