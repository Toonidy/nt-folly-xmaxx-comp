package graphql

import (
	"context"
	"fmt"
	"nt-folly-xmaxx-comp/internal/app/serve/dataloaders"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
	"nt-folly-xmaxx-comp/internal/pkg/utils"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

// Resolver contains the GQL resolvers.
type Resolver struct {
	Conn *pgxpool.Pool
	Log  *zap.Logger
}

// getTimeRangeRounded will round of the dates between the nearest X:X1 minute.
func getTimeRangeRounded(timeRange *gqlmodels.TimeRangeInput) (*gqlmodels.TimeRangeInput, error) {
	if timeRange == nil {
		return nil, nil
	}
	timeFrom := utils.TimeRound(timeRange.TimeFrom)
	timeTo := utils.TimeRound(timeRange.TimeTo)

	if timeFrom.After(timeTo) || timeFrom.Equal(timeTo) {
		return nil, fmt.Errorf("time range from value is invalid")
	}
	output := &gqlmodels.TimeRangeInput{
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	}
	return output, nil
}

////////////
//  User  //
////////////

type userResolver struct{ *Resolver }

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

func (r *userResolver) TotalPoints(ctx context.Context, obj *gqlmodels.User) (int, error) {
	pointLoader := dataloaders.GetLoadersFromContext(ctx).UserTotalPointsByID
	output, err := pointLoader.Load(obj.ID)
	if err != nil {
		return 0, fmt.Errorf("totalPoints dataloader failed: %w", err)
	}
	return output, nil
}

/////////////
//  Query  //
/////////////

type queryResolver struct{ *Resolver }

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Users is a query resolver that fetches all playing users.
func (r *queryResolver) Users(ctx context.Context) ([]*gqlmodels.User, error) {
	output := []*gqlmodels.User{}
	q := `
		SELECT u.id, u.username, u.display_name, u.membership_type, u.status, u.created_at, u.updated_at
		FROM users u
		WHERE u.deleted_at IS NULL`
	rows, err := r.Conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := gqlmodels.User{}
		err := rows.Scan(&row.ID, &row.Username, &row.DisplayName, &row.MembershipType, &row.Status, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to collect users: %w", err)
		}
		if row.DisplayName == "" {
			row.DisplayName = row.Username
		}
		output = append(output, &row)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("unable to collect users: %w", err)
	}
	return output, nil
}

// Competitions is a query resolver that fetches all available competitions.
func (r *queryResolver) Competitions(ctx context.Context, timeRange *gqlmodels.TimeRangeInput) ([]*gqlmodels.Competition, error) {
	timeRange, err := getTimeRangeRounded(timeRange)
	if err != nil {
		return nil, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Invalid time range received",
			Extensions: map[string]interface{}{
				"code": "INVALID_TIMERANGE",
			},
		}
	}
	output := []*gqlmodels.Competition{}
	args := []interface{}{}
	q := `
		SELECT c.id, c.status, c.multiplier, c.grind_rewards, c.point_rewards, c.speed_rewards, c.accuracy_rewards, c.from_at, c.to_at, c.updated_at
		FROM competitions c`
	if timeRange != nil {
		q += ` WHERE from_at >= $1 AND to_at <= $2`
		args = append(args, timeRange.TimeFrom, timeRange.TimeTo)
	}
	q += ` ORDER BY c.from_at ASC NULLS FIRST`
	rows, err := r.Conn.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to query competitions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := gqlmodels.Competition{
			GrindRewards:    []*gqlmodels.CompetitionPrize{},
			PointRewards:    []*gqlmodels.CompetitionPrize{},
			SpeedRewards:    []*gqlmodels.CompetitionPrize{},
			AccuracyRewards: []*gqlmodels.CompetitionPrize{},
		}
		grindRewards := []int{}
		pointRewards := []int{}
		speedRewards := []int{}
		accuracyRewards := []int{}
		err := rows.Scan(&row.ID, &row.Status, &row.Multiplier, &grindRewards, &pointRewards, &speedRewards, &accuracyRewards, &row.StartAt, &row.FinishAt, &row.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to collect competitions: %w", err)
		}
		for i, r := range grindRewards {
			prizeData := &gqlmodels.CompetitionPrize{
				Rank:   i + 1,
				Points: r,
			}
			row.GrindRewards = append(row.GrindRewards, prizeData)
		}
		for i, r := range pointRewards {
			prizeData := &gqlmodels.CompetitionPrize{
				Rank:   i + 1,
				Points: r,
			}
			row.PointRewards = append(row.PointRewards, prizeData)
		}
		for i, r := range speedRewards {
			prizeData := &gqlmodels.CompetitionPrize{
				Rank:   i + 1,
				Points: r,
			}
			row.SpeedRewards = append(row.SpeedRewards, prizeData)
		}
		for i, r := range accuracyRewards {
			prizeData := &gqlmodels.CompetitionPrize{
				Rank:   i + 1,
				Points: r,
			}
			row.AccuracyRewards = append(row.AccuracyRewards, prizeData)
		}
		output = append(output, &row)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("unable to collect competitions: %w", err)
	}
	return output, nil
}
