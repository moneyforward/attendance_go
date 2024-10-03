package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/walkure/slashbot_sample/handler"
)

func createSlackSocketClient() (*slack.Client, *socketmode.Client, error) {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		return nil, nil, errors.New("SLACK_APP_TOKEN must be set")
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		return nil, nil, errors.New("SLACK_APP_TOKEN must have the prefix \"xapp-\"")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, nil, errors.New("SLACK_BOT_TOKEN must be set")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, nil, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
	}

	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		//slack.OptionDebug(true),
		//slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)

	client := socketmode.New(
		api,
		//socketmode.OptionDebug(true),
		//socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return api, client, nil
}

func main() {

	//slog.SetLogLoggerLevel(slog.LevelDebug)

	api, client, err := createSlackSocketClient()
	if err != nil {
		slog.Error("failure to connect to Slack", slog.String("error", err.Error()))
		os.Exit(-1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				slog.InfoContext(ctx, "connecting to Slack with Socket Mode")
			case socketmode.EventTypeHello:
				slog.InfoContext(ctx, "hello from Slack with Socket Mode")
			case socketmode.EventTypeConnectionError:
				slog.ErrorContext(ctx, "connection failed. Retry later", slog.String("error", fmt.Sprintf("%+v", evt.Data)))
			case socketmode.EventTypeConnected:
				slog.InfoContext(ctx, "connected to Slack with Socket Mode")
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					slog.WarnContext(ctx, "failed to parse slash command")
					continue
				}
				msg, err := handler.HandleSlashCommandEvent(ctx, api, cmd)
				if err != nil {
					slog.ErrorContext(ctx, "failed to handle slash command", slog.String("error", err.Error()))
					msg = nil
				}
				slog.DebugContext(ctx, "sending response", "msg", msg)
				client.Ack(*evt.Request, msg)

			default:
				slog.WarnContext(ctx, "unexpected event type received", slog.String("type", string(evt.Type)))
			}
		}
	}()

	slog.Error("loop exit", slog.String("error", client.Run().Error()))
}
