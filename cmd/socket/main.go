package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sh := socketmode.NewSocketmodeHandler(client)

	sh.Handle(socketmode.EventTypeConnecting, func(e *socketmode.Event, c *socketmode.Client) {
		slog.Info("connecting to Slack with Socket Mode")
	})

	sh.Handle(socketmode.EventTypeHello, func(e *socketmode.Event, c *socketmode.Client) {
		slog.Info("hello from Slack with Socket Mode")
	})

	sh.Handle(socketmode.EventTypeConnectionError, func(e *socketmode.Event, c *socketmode.Client) {
		slog.Error("connection failed. Retry later", slog.String("error", fmt.Sprintf("%+v", e.Data)))
	})

	sh.Handle(socketmode.EventTypeConnected, func(e *socketmode.Event, c *socketmode.Client) {
		slog.Info("connected to Slack with Socket Mode")
	})

	sh.Handle(socketmode.EventTypeSlashCommand, func(e *socketmode.Event, c *socketmode.Client) {
		slog.Debug("slash command received", "command", e.Data)
		cmd, ok := e.Data.(slack.SlashCommand)
		if !ok {
			slog.Warn("failed to parse slash command")
			return
		}
		msg, err := handler.HandleSlashCommandEvent(ctx, api, cmd)
		if err != nil {
			slog.Error("failed to handle slash command", slog.String("error", err.Error()))
			msg = nil
		}
		slog.Debug("sending response", "msg", msg)
		c.Ack(*e.Request, msg)
	})

	slog.Error("loop exit", slog.String("error", sh.RunEventLoopContext(ctx).Error()))
}
