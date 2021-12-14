package graphql

import (
	"context"
	"fmt"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
	"nt-folly-xmaxx-comp/internal/pkg/utils"

	"github.com/jackc/pgx/v4/pgxpool"
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

/////////////
//  Query  //
/////////////

type queryResolver struct{ *Resolver }

// Query points query func from graphql_schema_gen.go
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Users is a query resolver that fetches all playing users.
func (r *queryResolver) Users(ctx context.Context, timeRange *gqlmodels.TimeRangeInput) ([]*gqlmodels.User, error) {
	output := []*gqlmodels.User{}
	q := `
		SELECT id, username, display_name, membership_type, status, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL`
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
	output := []*gqlmodels.Competition{}
	return output, nil
}
