package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/trevex/zanzigo/server"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mlog := log.WithGroup("main")

	undo, err := maxprocs.Set(maxprocs.Logger(func(format string, a ...any) {
		log.Info(fmt.Sprintf(format, a...))
	}))
	defer undo()
	if err != nil {
		mlog.Error("failed to set GOMAXPROCS: %v", slog.Any("error", err))
		os.Exit(-2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd := &cobra.Command{
		Use: "zanzigo action [flags]",
		// TODO: properly document
	}
	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)

	// Add all sub-commands
	rootCmd.AddCommand(server.NewServerCmd(log.WithGroup("server")))

	// Make sure to cancel the context if a signal was received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		mlog.Info("received signal", slog.String("signal", sig.String()))
		cancel()
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		mlog.Error("command failed", slog.Any("error", err))
		os.Exit(-1)
	}
}
