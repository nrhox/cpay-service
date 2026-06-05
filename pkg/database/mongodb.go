package database

import (
	"context"
	"log"
	"os"

	"github.com/nrhox/cpay-service/internal/config"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DisconnectFunc func(context.Context) error

func NewMongoConnetion(cfg config.Config) (db *mongo.Database, disconnect DisconnectFunc, err error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(cfg.Mongo.DbUrl).SetServerAPIOptions(serverAPI)

	if cfg.Mode == config.MODE_DEBUG {
		logger := log.New(os.Stdout, "[MONGO-QUERY] ", log.LstdFlags)

		cmdMonitor := &event.CommandMonitor{
			Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
				cmdString := evt.Command.String()

				logger.Printf("[QUERY] %s\n", cmdString)
			},

			Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
				logger.Printf("[SUCCESS] %s Duration: %v\n", evt.CommandName, evt.Duration)
			},

			Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
				logger.Printf("[ERROR] %v\n", evt.Failure)
			},
		}

		opts.SetMonitor(cmdMonitor)
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, nil, err
	}

	return client.Database(cfg.Mongo.DatabaseName), func(ctx context.Context) error {
		return client.Disconnect(ctx)
	}, nil
}
