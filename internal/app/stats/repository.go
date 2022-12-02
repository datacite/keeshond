package stats

import (
	"time"

	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/datacite/keeshond/internal/app/event"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatsRepositoryReader interface {
	// For a specific repository return aggregates, for the specified time query
	Aggregate(repoId string, query Query) AggregateResult
	// For a specific repository return list of time series results, for the specified time query
	Timeseries(repoId string, query Query) []TimeseriesResult
	// For a specific repository return for the specified time query and grouped by PID.
	BreakdownByPID(repoId string, query Query, limit int, page int) []BreakdownResult
	// Return count of unique PIDs for repository over time period.
	CountUniquePID(repoId string, query Query) int64
}

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{
		db: db,
	}
}

func (repository *StatsRepository) Aggregate(repoId string, query Query) AggregateResult {
	var result AggregateResult

	// Get timestamp scope from query start and end
	timestampScope := TimestampCustom(query.Start, query.End)

	repository.db.
		Clauses(
			exclause.NewWith(
				"time_period_deduped", repository.db.Model(&event.Event{}).
					Select("name, pid, session_id, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
					Scopes(RepoId(repoId), timestampScope).
					Group("name, pid, session_id, interval_alias order by interval_alias"),
			),
		).Table("time_period_deduped").
		Select("countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads").
		Scan(&result)

	return result
}

func (repository *StatsRepository) Timeseries(repoId string, query Query) []TimeseriesResult {
	var result []TimeseriesResult

	// Get timestamp scope from query start and end
	timestampScope := TimestampCustom(query.Start, query.End)

	db := repository.db.Debug().
		Clauses(
			exclause.NewWith(
				"time_period_deduped", repository.db.Model(&event.Event{}).
					Select("name, pid, session_id, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
					Scopes(RepoId(repoId), timestampScope).
					Group("name, pid, session_id, interval_alias order by interval_alias"),
			),
		).Table("time_period_deduped")

	switch query.Interval {
	case "month":
		db = db.Select("toStartOfMonth(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
	case "hour":
		db = db.Select("toStartOfHour(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
	case "day":
		fallthrough
	default:
		db = db.Select("toStartOfDay(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
	}

	db = db.Group("date")

	switch query.Interval {
	case "month":
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfMonth(?) TO toStartOfMonth(?) STEP INTERVAL 1 MONTH", Vars: []interface{}{query.Start, query.End.AddDate(0, 1, 0)}, WithoutParentheses: true},
		})
	case "hour":
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfHour(?) TO toStartOfHour(?) STEP INTERVAL 1 HOUR", Vars: []interface{}{query.Start, query.End}, WithoutParentheses: true},
		})
	case "day":
		fallthrough
	default:
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfDay(?) TO toStartOfDay(?) STEP INTERVAL 1 DAY", Vars: []interface{}{query.Start, query.End}, WithoutParentheses: true},
		})
	}

	db.Scan(&result)

	return result
}

func (repository *StatsRepository) BreakdownByPID(repoId string, query Query, page int, pageSize int) []BreakdownResult {
	var result []BreakdownResult

	// Get timestamp scope from query start and end
	timestampScope := TimestampCustom(query.Start, query.End)

	repository.db.
		Clauses(
			exclause.NewWith(
				"time_period_deduped", repository.db.Model(&event.Event{}).
					Select("name, pid, session_id, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
					Scopes(RepoId(repoId), timestampScope).
					Group("name, pid, session_id, interval_alias order by interval_alias"),
			),
		).Table("time_period_deduped").
		Select("pid, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads").
		Group("pid").
		Scopes(Paginate(page, pageSize)).
		Scan(&result)

	return result
}

func (repository *StatsRepository) CountUniquePID(repoId string, query Query) int64 {
	var count int64

	// Get timestamp scope from query start and end
	timestampScope := TimestampCustom(query.Start, query.End)

	repository.db.Model(&event.Event{}).Distinct("pid").Count(&count).Scopes(RepoId(repoId), timestampScope)

	return count
}

// Following are scopes for the event model that can be used
// to build up queries dynamically.

func NameAndRepoId(name string, repo_id string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name).Where("repo_id = ?", repo_id)
	}
}

func RepoId(repo_id string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("repo_id = ?", repo_id)
	}
}

func PID(pid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("pid = ?", pid)
	}
}

func SelectDateByDay(db *gorm.DB) *gorm.DB {
	return db.Select("toStartOfDay(interval_alias) as date")
}

func TimestampCustom(start_date time.Time, end_date time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ?", start_date).Where("timestamp < ?", end_date)
	}
}

func Paginate(page int, page_size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		if page_size == 0 {
			page_size = 100
		}

		offset := (page - 1) * page_size
		return db.Offset(offset).Limit(page_size)
	}
}
