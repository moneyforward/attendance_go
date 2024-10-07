package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/slack-go/slack"
	"github.com/walkure/slashbot_sample/handler"
	"github.com/walkure/slashbot_sample/pkg/logger"
)

func createSlackClient() (*slack.Client, error) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, errors.New("SLACK_BOT_TOKEN must be set")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
	}

	return slack.New(
		botToken,
		//slack.OptionDebug(true),
		//slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	), nil
}

func main() {

	slog.SetDefault(logger.InitalizeLogger())

	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	api, err := createSlackClient()
	if err != nil {
		slog.Error("failure to create slack client", slog.String("error", err.Error()))
		return
	}

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.WarnContext(ctx, "failed to create secrets verifier", slog.String("error", err.Error()))
			return
		}

		r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
		cmd, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to parse slash command", slog.String("error", err.Error()))
			return
		}

		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			slog.WarnContext(ctx, "failed to verify request", slog.String("error", err.Error()))
			return
		}

		msg, err := handler.HandleSlashCommandEvent(ctx, api, cmd)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to handle slash command", slog.String("error", err.Error()))
			return
		}

		b, err := json.Marshal(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to marshal response", slog.String("error", err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

		slog.DebugContext(ctx, "response sent", "msg", msg)

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serv := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server listening", slog.String("port", port))

		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server shutdown", slog.String("error", err.Error()))
		}
	}()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	slog.Warn("shutting down server")
	serv.Shutdown(ctx)

}
