package handler

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

func HandleSlashCommandEvent(ctx context.Context, api *slack.Client, cmd slack.SlashCommand) (*slack.Msg, error) {
	return &slack.Msg{
		Text:         fmt.Sprint("Hello, ", cmd.UserName, " cmd", cmd.Command, " ", cmd.Text),
		ResponseType: slack.ResponseTypeInChannel,
	}, nil
}
