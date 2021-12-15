package dataloaders

import (
	"context"
	"fmt"
	"net/http"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4/pgxpool"
)

type contextKey string

const key = contextKey("dataloaders")

// Loaders hold references to the individual dataloaders.
type Loaders struct {
	UserTotalPointsByID *UserTotalPointsLoader
}

// newLoaderse initializes individual loaders.
func newLoaders(ctx context.Context, conn *pgxpool.Pool) *Loaders {
	return &Loaders{
		UserTotalPointsByID: userTotalPointLoader(conn),
	}
}

// GetLoadersFromContext retrives dataloaders from context
func GetLoadersFromContext(ctx context.Context) *Loaders {
	return ctx.Value(key).(*Loaders)
}

// Middleware stores Loaders as a requested-scored context value.
func Middleware(conn *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			loaders := newLoaders(ctx, conn)
			augmentedCtx := context.WithValue(ctx, key, loaders)
			r = r.WithContext(augmentedCtx)
			next.ServeHTTP(w, r)
		})
	}
}

///////////////////
//  Dataloaders  //
///////////////////

// userTotalPointLoader fetches the total points for the following resolver:
// * user -> totalPoints
func userTotalPointLoader(conn *pgxpool.Pool) *UserTotalPointsLoader {
	type userPointResult struct {
		userID string
		points int
	}
	return NewUserTotalPointsLoader(
		UserTotalPointsLoaderConfig{
			Fetch: func(ids []string) ([]int, []error) {
				if len(ids) == 0 {
					return []int{}, nil
				}

				// Query users
				q, args, err := db.QueryBuilder.
					Select(
						goqu.C("user_id"),
						goqu.L("SUM(grind_reward + accuracy_reward + speed_reward + point_reward)"),
					).
					From("competition_results").
					Where(
						goqu.Ex{
							"user_id": ids,
						},
					).
					GroupBy(goqu.C("user_id")).
					ToSQL()
				if err != nil {
					return nil, []error{fmt.Errorf("failed to build user total points query: %w", err)}
				}
				results := []userPointResult{}
				rows, err := conn.Query(context.Background(), q, args...)
				if err != nil {
					return nil, []error{fmt.Errorf("failed to query user total points: %w", err)}
				}
				defer rows.Close()
				for rows.Next() {
					var row userPointResult
					err := rows.Scan(&row.userID, &row.points)
					if err != nil {
						return nil, []error{fmt.Errorf("failed to scan user total points: %w", err)}
					}
					results = append(results, row)
				}
				err = rows.Err()
				if err != nil {
					return nil, []error{fmt.Errorf("an error occurred while scanning user total points: %w", err)}
				}

				// Generate output
				output := []int{}
				for _, key := range ids {
					for _, row := range results {
						if row.userID == key {
							output = append(output, row.points)
							break
						}
					}
				}
				return output, nil
			},
			Wait:     1 * time.Millisecond,
			MaxBatch: 100,
		},
	)
}
