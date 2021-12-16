package cli

import (
	"nt-folly-xmaxx-comp/internal/app/migrate/seed"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// dbSeedCompetition represents the seed-comp command
var dbSeedCompetition = &cobra.Command{
	Use:   "seed-comp",
	Short: "runs db migrate drop command.",
	Long:  "Runs DB Migration Drop command.",
	Run: func(cmd *cobra.Command, args []string) {
		timeFromValue, err := cmd.Flags().GetString("time_from")
		if err != nil {
			logger.Error("unable to read time_from flag", zap.Error(err))
			return
		}
		timeToValue, err := cmd.Flags().GetString("time_to")
		if err != nil {
			logger.Error("unable to read time_to flag", zap.Error(err))
			return
		}
		timeFrom, err := time.Parse(time.RFC3339, timeFromValue)
		if err != nil {
			logger.Error("unable to parse time_from flag", zap.Error(err))
			return
		}
		timeTo, err := time.Parse(time.RFC3339, timeToValue)
		if err != nil {
			logger.Error("unable to parse time_to flag", zap.Error(err))
			return
		}
		ctx := cmd.Context()
		conn, err := db.ConnectPool(ctx, db.GetConnectionString(), logger)
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}
		logger.Info("db seed comp started")
		err = seed.SetupCompetition(ctx, conn, timeFrom, timeTo)
		if err != nil {
			logger.Error("failed to db seed comp", zap.Error(err))
			return
		}
		logger.Info("db seed comp finished")
	},
}

func init() {
	dbSeedCompetition.Flags().String("time_from", "", "comp time from (it'll round down to the nearest 1st minute)")
	dbSeedCompetition.Flags().String("time_to", "", "comp time to (it'll round down to the nearest 1st minute)")

	rootCmd.AddCommand(dbSeedCompetition)
}
