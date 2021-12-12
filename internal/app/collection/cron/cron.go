package cron

import (
	"context"
	"encoding/json"
	"errors"
	"nt-folly-xmaxx-comp/internal/pkg/utils"
	"nt-folly-xmaxx-comp/pkg/nitrotype"

	"github.com/go-logr/zapr"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// NewCronService creates a new cron service ready to be activated
func NewCronService(ctx context.Context, conn *pgxpool.Pool, log *zap.Logger, apiClient nitrotype.APIClient, teamTag string, teamID int) *cron.Cron {
	logger := zapr.NewLogger(log)
	c := cron.New(
		cron.WithChain(cron.DelayIfStillRunning(logger)),
	)
	c.AddFunc("1,11,21,31,41,51 * * * *", syncTeams(ctx, conn, log, apiClient, teamTag, teamID))
	return c
}

// syncTeams is the scheduled task function that collect Nitro Type Team Logs.
func syncTeams(ctx context.Context, conn *pgxpool.Pool, l *zap.Logger, apiClient nitrotype.APIClient, teamTag string, teamID int) func() {
	log := l.With(
		zap.String("job", "syncTeams"),
		zap.String("team", teamTag),
	)

	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("recovering from panic", zap.Any("panic", r))
			}
		}()

		log.Info("sync teams started")

		// Get Previous Log
		var (
			prevRequestID pgtype.UUID
			prevLogID     pgtype.UUID
		)
		q := `
			SELECT id, api_team_log_id
			FROM nt_api_team_log_requests
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1`
		err := conn.QueryRow(ctx, q).Scan(&prevRequestID, &prevLogID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error("unable to query previous log", zap.Error(err))
			return
		}

		// Grab Latest Stats
		teamData, err := apiClient.GetTeam(teamTag)
		if err != nil || !teamData.Success || teamData.Data.Info == nil {
			log.Error("unable to pull team log", zap.Error(err))

			// Record Fail Request
			if prevLogID.Status == pgtype.Present && prevRequestID.Status == pgtype.Present {
				responseType := "ERROR"
				description := "Unknown error"
				if err != nil {
					description = err.Error()
				} else if !teamData.Success || teamData.Data.Info == nil {
					description = "Team API Request Failed"
				}
				q = `
					INSERT INTO nt_api_team_log_requests (prev_id, api_team_log_id, response_type, description)
					VALUES ($1, $2, $3, $4)`
				_, err = conn.Exec(ctx, q, prevRequestID, prevLogID, responseType, description)
				if err != nil {
					log.Error("unable to insert request log (error)", zap.Error(err))
				}
			}
			return
		}

		// Check if data doesn't matches team
		if teamID != teamData.Data.Info.TeamID {
			log.Error("team has changed", zap.Int("teamID", teamData.Data.Info.TeamID))
			return
		}

		// Calculate Hash
		data, err := json.Marshal(teamData)
		if err != nil {
			log.Error("unable to marshal team data", zap.Error(err))
			return
		}
		hash, err := utils.HashData(data)
		if err != nil {
			log.Error("unable to calculate team data hash", zap.Error(err))
			return
		}

		// Insert Team Log
		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Error("unable to start recording team data", zap.Error(err))
			return
		}

		logID := ""
		responseType := "NEW"
		description := "New log download"
		q = `SELECT id FROM nt_api_team_logs WHERE hash = $1`
		err = tx.QueryRow(ctx, q, hash).Scan(&logID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error("unable to find existing team log", zap.Error(err))
			return
		}
		if logID == "" {
			q := `
				INSERT INTO nt_api_team_logs (hash, log_data)
				VALUES ($1, $2)
				ON CONFLICT (hash) DO NOTHING
				RETURNING id`
			err = tx.QueryRow(ctx, q, hash, data).Scan(&logID)
			if err != nil {
				log.Error("unable to insert team log", zap.Error(err))
				return
			}
		}
		if logID == "" {
			log.Error("unable to find team log id (blank data)")
			return
		}

		if prevLogID.Status == pgtype.Present {
			prevLogIDText := ""
			prevLogID.AssignTo(&prevLogIDText)
			if prevLogIDText == logID {
				responseType = "CACHE"
				description = "Same log found"
			}
		}
		if prevRequestID.Status != pgtype.Present {
			prevRequestID.Set(nil)
		}

		// Insert Team Log Request
		q = `
			INSERT INTO nt_api_team_log_requests (prev_id, api_team_log_id, response_type, description)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(ctx, q, prevRequestID, logID, responseType, description)
		if err != nil {
			log.Error("unable to insert team log request", zap.Error(err))
			return
		}

		// Commit Transaction
		err = tx.Commit(ctx)
		if err != nil {
			log.Error("unable to fininsh recording team data", zap.Error(err))
			return
		}

		log.Info("sync teams completed")
	}
}
