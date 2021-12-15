package graphql

import (
	"context"
	"fmt"
	"nt-folly-xmaxx-comp/internal/app/serve/dataloaders"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
	"nt-folly-xmaxx-comp/internal/pkg/utils"
	"time"

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
func (r *queryResolver) Users(ctx context.Context, timeRange *gqlmodels.TimeRangeInput) ([]*gqlmodels.User, error) {
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
	output := []*gqlmodels.User{}
	args := []interface{}{}
	q := `
		SELECT u.id, u.username, u.display_name, u.membership_type, u.status, u.created_at, u.updated_at
		FROM users u
		WHERE u.deleted_at IS NULL`
	if timeRange != nil {
		q += ` AND EXISTS (
			SELECT 1
			FROM user_records _ur
			WHERE _ur.user_id = u.id
				AND _ur.created_at BETWEEN $1 AND $2
			LIMIT 1
		)`
		args = append(args, timeRange.TimeFrom, timeRange.TimeTo.Add(5*time.Minute))
	}

	rows, err := r.Conn.Query(ctx, q, args...)
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
