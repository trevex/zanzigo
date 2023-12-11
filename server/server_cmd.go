package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/api/zanzigo/v1/zanzigov1connect"
	"github.com/trevex/zanzigo/storage/postgres"
	"github.com/trevex/zanzigo/storage/sqlite3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewServerCmd(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use: "server [flags] [model-file]",
		// TODO: properly document
	}

	var (
		port          int
		sqliteFile    string
		postgresURL   string
		useFunctions  bool
		runMigrations bool
		maxDepth      int
	)

	flags := cmd.Flags()
	flags.IntVar(&port, "port", 4000, "port the server is listening on")
	flags.StringVar(&sqliteFile, "sqlite-file", "./zanzigo.db", "sqlite database file (will be ignored if --postgres-url is set)")
	flags.StringVar(&postgresURL, "postgres-url", "", "postgres database to connect to")
	flags.BoolVar(&useFunctions, "use-functions", false, "postgres-specific flag enable the use of function to run checks via functions")
	flags.BoolVar(&runMigrations, "run-migrations", true, "run database migrations on the configured database")
	flags.IntVar(&maxDepth, "max-depth", 16, "maximum depth to traverse relationships")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if len(args) != 1 {
			return fmt.Errorf("model-file required as first argument")
		}

		model, err := loadModel(args[0])
		if err != nil {
			return err
		}

		if runMigrations {
			if postgresURL != "" {
				err = postgres.RunMigrations(postgresURL)
			} else {
				err = sqlite3.RunMigrations(sqliteFile)
			}
			if err != nil {
				return err
			}
		}

		var storage zanzigo.Storage
		if postgresURL != "" {
			options := []postgres.PostgresOption{}
			if useFunctions {
				options = append(options, postgres.UseFunctions())
			}
			storage, err = postgres.NewPostgresStorage(postgresURL, options...)
		} else {
			storage, err = sqlite3.NewSQLite3Storage(sqliteFile)
		}
		if err != nil {
			return err
		}

		resolver, err := zanzigo.NewResolver(model, storage, maxDepth)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.Handle(zanzigov1connect.NewZanzigoServiceHandler(NewZanzigoServiceHandler(log.WithGroup("handler"), model, storage, resolver)))
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

func loadModel(filename string) (*zanzigo.Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	objects := zanzigo.ObjectMap{}
	err = json.Unmarshal(data, &objects)
	if err != nil {
		return nil, err
	}

	return zanzigo.NewModel(objects)
}
