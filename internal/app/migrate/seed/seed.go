package seed

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrAlreadySeeded = fmt.Errorf("data already seeded")
	DefaultRewards   = []int{10, 7, 5, 3, 1}
)

// SetupCompetition seeds in the competition data.
func SetupCompetition(ctx context.Context, conn *pgxpool.Pool, timeFrom time.Time, timeTo time.Time) error {
	count := 0
	q := `SELECT COUNT(*) FROM competitions`
	err := conn.QueryRow(ctx, q).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if already seeded: %w", err)
	}
	if count > 0 {
		return ErrAlreadySeeded
	}

	timeFrom = time.Date(timeFrom.Year(), timeFrom.Month(), timeFrom.Day(), timeFrom.Hour(), int(math.Floor(float64(timeFrom.Minute())/10)*10)+1, 0, 0, timeFrom.Location())
	timeTo = time.Date(timeTo.Year(), timeTo.Month(), timeTo.Day(), timeTo.Hour(), int(math.Floor(float64(timeTo.Minute())/10)*10)+1, 0, 0, timeTo.Location())

	batch := &pgx.Batch{}
	for {
		fromAt := timeFrom
		toAt := timeFrom.Add(time.Minute * 10)

		multiplier := 1
		randNumber := rand.Intn(100)

		if randNumber <= 5 {
			multiplier = 8
		} else if randNumber <= 10 {
			multiplier = 4
		} else if randNumber <= 20 {
			multiplier = 2
		}

		q := `
			INSERT INTO competitions (multiplier, grind_rewards, point_rewards, speed_rewards, accuracy_rewards, from_at, to_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
		batch.Queue(q, multiplier, DefaultRewards, DefaultRewards, DefaultRewards, DefaultRewards, fromAt, toAt)

		timeFrom = toAt
		if timeFrom.Equal(timeTo) || timeFrom.After(timeTo) {
			break
		}
	}

	batchRequest := conn.SendBatch(ctx, batch)
	defer batchRequest.Close()
	for i := 0; i < batch.Len(); i++ {
		_, err = batchRequest.Exec()
		if err != nil {
			return fmt.Errorf("failed to seed the database: %w", err)
		}
	}
	return nil
}
