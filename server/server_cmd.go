package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/trevex/zanzigo/api/zanzigo/v1/zanzigov1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewServerCmd(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server [flags] [model-file]",
		Short: "TODO",
	}

	var (
		port int
	)

	flags := cmd.Flags()
	flags.IntVar(&port, "port", 4000, "port the server is listening on")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		mux := http.NewServeMux()
		// TODO: model, storage, resolver
		mux.Handle(zanzigov1connect.NewZanzigoServiceHandler(NewZanzigoServiceHandler(log.WithGroup("handler"), nil, nil, nil)))

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: h2c.NewHandler(mux, &http2.Server{}),
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		}

		// Start HTTP server at :4000.
		log.Info(fmt.Sprintf("started server on 0.0.0.0:%d, http://localhost:%d", port, port))
		go func() {
			err := server.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("server gracefully closed")
			} else if err != nil {
				log.Error("error listening on server", slog.Any("error", err))
			}
		}()

		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := server.Shutdown(ctxShutdown); err != nil {
			log.Error("error on server shutdown", slog.Any("error", err))
			return err
		}
		return nil
	}

	return cmd
}
