package dataloaders

import (
	"context"
	"fmt"
	"net/http"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4/pgxpool"
)

type contextKey string

const key = contextKey("dataloaders")

// Loaders hold references to the individual dataloaders.
type Loaders struct {
	UserTotalPointsByID        *UserTotalPointsLoader
	CompetitionLeaderboardByID *CompetitionLeaderboardLoader
}

// newLoaderse initializes individual loaders.
func newLoaders(ctx context.Context, conn *pgxpool.Pool) *Loaders {
	return &Loaders{
		UserTotalPointsByID:        userTotalPointLoader(conn),
		CompetitionLeaderboardByID: competitionLeaderboardLoader(conn),
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

// competitionLeaderboardLoader fetches the leaderboard result for the following resolver:
// * competition -> leaderboard
func competitionLeaderboardLoader(conn *pgxpool.Pool) *CompetitionLeaderboardLoader {
	type competitionLeaderboardResult struct {
		competitionID  string
		userID         string
		grind          int
		grindRank      int
		speed          float64
		speedRank      int
		point          int
		pointRank      int
		accuracy       float64
		accuracyRank   int
		username       string
		displayName    string
		membershipType string
		status         string
		createdAt      time.Time
		updatedAt      time.Time
	}
	return NewCompetitionLeaderboardLoader(
		CompetitionLeaderboardLoaderConfig{
			Fetch: func(ids []string) ([][]*gqlmodels.CompetitionUser, []error) {
				if len(ids) == 0 {
					return [][]*gqlmodels.CompetitionUser{}, nil
				}

				// Query leaderboard data
				q, args, err := db.QueryBuilder.
					Select(
						goqu.C("r.competition_id"),
						goqu.C("r.user_id"),
						goqu.C("r.grind"),
						goqu.C("r.grind_rank"),
						goqu.C("r.accuracy"),
						goqu.C("r.accuracy_rank"),
						goqu.C("r.speed"),
						goqu.C("r.speed_rank"),
						goqu.C("r.point"),
						goqu.C("r.point_rank"),
						goqu.C("u.username"),
						goqu.C("u.display_name"),
						goqu.C("u.membership_type"),
						goqu.C("u.status"),
						goqu.C("u.created_at"),
						goqu.C("u.updated_at"),
					).
					From("competition_results r").
					InnerJoin(
						goqu.T("users u"),
						goqu.On(goqu.L("u.id = r.user_id")),
					).
					Where(goqu.Ex{
						"competition_id": ids,
					}).
					ToSQL()
				if err != nil {
					return nil, []error{fmt.Errorf("failed to build competition leaderboard: %w", err)}
				}
				rows, err := conn.Query(context.Background(), q, args...)
				if err != nil {
					return nil, []error{fmt.Errorf("failed to query competition leaderboard: %w", err)}
				}
				defer rows.Close()
				results := []competitionLeaderboardResult{}
				for rows.Next() {
					var row competitionLeaderboardResult
					err := rows.Scan(
						&row.competitionID, &row.userID, &row.grind, &row.grindRank, &row.accuracy,
						&row.accuracyRank, &row.speed, &row.speedRank, &row.point, &row.pointRank,
						&row.username, &row.displayName, &row.membershipType, &row.status, &row.createdAt, &row.updatedAt,
					)
					if err != nil {
						return nil, []error{fmt.Errorf("failed to scan user total points: %w", err)}
					}
					results = append(results, row)
				}
				err = rows.Err()
				if err != nil {
					return nil, []error{fmt.Errorf("an error occurred while scanning user total points: %w", err)}
				}

				// Return output
				output := [][]*gqlmodels.CompetitionUser{}
				for _, key := range ids {
					rows := []*gqlmodels.CompetitionUser{}
					for _, row := range results {
						if row.userID == key {
							if row.displayName == "" {
								row.displayName = row.username
							}
							rows = append(rows, &gqlmodels.CompetitionUser{
								ID:            fmt.Sprintf("%s::%s", row.competitionID, row.userID),
								GrindScore:    row.grind,
								GrindRank:     row.grindRank,
								SpeedScore:    row.speed,
								SpeedRank:     row.speedRank,
								PointScore:    row.point,
								PointRank:     row.pointRank,
								AccuracyScore: row.accuracy,
								AccuracyRank:  row.accuracyRank,
								User: &gqlmodels.User{
									ID:             row.userID,
									Username:       row.username,
									DisplayName:    row.displayName,
									MembershipType: gqlmodels.MembershipType(row.membershipType),
									Status:         gqlmodels.UserStatus(row.status),
									CreatedAt:      row.createdAt,
									UpdatedAt:      row.updatedAt,
								},
							})
						}
					}
					output = append(output, rows)
				}
				return output, nil
			},
			Wait:     1 * time.Millisecond,
			MaxBatch: 100,
		},
	)
}

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
