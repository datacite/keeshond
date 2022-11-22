package stats

import (
	"fmt"
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
}

type StatsRepository struct {
	db 		*gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{
		db: db,
	}
}

// func BaseQuery(repoId string, db *gorm.DB) *gorm.DB {
// 	return db.
// 	Clauses(
// 		exclause.NewWith(
// 			"time_period_deduped", db.Model(&event.Event{}).
// 			Select("name, pid, session_id, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
// 			Scopes(RepoId(repoId)).
// 			Group("name, pid, session_id, interval_alias order by interval_alias"),
// 		),
// 	).Table("time_period_deduped").
// 	Select("countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads").
// 	Group("name, pid").
// 	Limit(10)
// }


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

	db := repository.db.
	Clauses(
		exclause.NewWith(
			"time_period_deduped", repository.db.Model(&event.Event{}).
			Select("name, pid, session_id, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(RepoId(repoId), timestampScope).
			Group("name, pid, session_id, interval_alias order by interval_alias"),
		),
	).Table("time_period_deduped")

	switch query.Interval {
		case "day":
		db = db.Select("toStartOfDay(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
		case "month":
		db = db.Select("toStartOfMonth(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
		case "hour":
		db = db.Select("toStartOfHour(interval_alias) as date, countIf(name = 'view') as total_views, uniqIf(session_id, name = 'view') as unique_views, countIf(name = 'download') as total_downloads, uniqIf(session_id, name = 'download') as unique_downloads")
	}

	db = db.Group("date")

	switch query.Interval {
		case "day":
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfDay(?) TO toStartOfDay(?) STEP INTERVAL 1 DAY", Vars: []interface{}{query.Start, query.End}, WithoutParentheses: true},
		})
		case "month":
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfMonth(?) TO toStartOfMonth(?) STEP INTERVAL 1 MONTH", Vars: []interface{}{query.Start, query.End}, WithoutParentheses: true},
		})
		case "hour":
		db = db.Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "date WITH FILL FROM toStartOfHour(?) TO toStartOfHour(?) STEP INTERVAL 1 HOUR", Vars: []interface{}{query.Start, query.End}, WithoutParentheses: true},
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

	// Debug print result
	fmt.Printf("%+v\n", result)

	return result
}

// // This query is for getting based on an interval the total unique metric events
// // We do an inner CTE to be able to remove duplicates based on the interval of 30 secs
// // We then can count the returning CTE table to get our total de duplicated events.
// // For COUNTER this is referred to as Double Click filtering.
// func (repository *StatsRepository) GetTotalForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat {
// 	var count int64
// 	repository.db.
// 	Clauses(
// 		exclause.NewWith(
// 			"interval_calc", repository.db.Model(&event.Event{}).
// 			Select("toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
// 			Scopes(NameAndRepoId(metricName, repoId), PID(pid), interval).
// 			Group("interval_alias"),
// 		),
// 	).Table("interval_calc").Count(&count)

// 	return PidStat {
// 		Metric: metricName,
// 		Pid: pid,
// 		Count: count,
// 	}
// }

// func (repository *StatsRepository) GetTotalInToday(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIsToday)
// }

// func (repository *StatsRepository) GetTotalInLast7Days(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn7Days)
// }

// func (repository *StatsRepository) GetTotalInLast30Days(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn30Days)
// }

// // Similiar to the query for specific pid, builds a CTE that is grouped by the interval of 30 secs to remove duplicates in interval
// // But also returns the pid and then we do a group by again to get counts for pids.
// func (repository *StatsRepository) GetTotalsByPidForInterval(metricName string, repoId string, interval func(*gorm.DB) *gorm.DB) []PidStat {
// 	var result []PidStat

// 	repository.db.Debug().
// 	Clauses(
// 		exclause.NewWith(
// 			"interval_calc", repository.db.Model(&event.Event{}).
// 			Select("name, pid, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
// 			Scopes(NameAndRepoId(metricName, repoId), interval).
// 			Group("name, pid, interval_alias"),
// 		),
// 	).Table("interval_calc").Select("name as metric, pid, count(*) as count").Group("name, pid").Limit(10).Scan(&result)

// 	return result
// }

// func (repository *StatsRepository) GetTotalsByPidInToday(metricName string, repo_id string) []PidStat {
// 	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIsToday)
// }

// func (repository *StatsRepository) GetTotalsByPidInLast7Days(metricName string, repo_id string) []PidStat {
// 	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn7Days)
// }

// func (repository *StatsRepository) GetTotalsByPidInLast30Days(metricName string, repo_id string) []PidStat {
// 	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn30Days)
// }

// // To get the unique values, this can be achieved with a simple distinct query across events for specific pids/repos
// // This doesn't need the deduping logic because sessions are always same within 1 hour, therefore any duplicated events
// // will not appear in the result count.
// // For COUNTER this is referred to as Counting Unique Datasets
// func (repository *StatsRepository) GetUniqueForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat {
// 	var count int64
// 	repository.db.Model(&event.Event{}).
// 		Distinct("session_id").
// 		Scopes(NameAndRepoId(metricName, repoId), PID(pid), interval).
// 		Count(&count)

// 	return PidStat {
// 		Metric: metricName,
// 		Pid: pid,
// 		Count: count,
// 	}
// }

// func (repository *StatsRepository) GetUniqueInToday(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIsToday)
// }

// func (repository *StatsRepository) GetUniqueInLast7Days(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIn7Days)
// }

// func (repository *StatsRepository) GetUniqueInLast30Days(metricName string, repo_id string, pid string) PidStat {
// 	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIn30Days)
// }

// // Similiar to the unique values by PID, we can do the same as this except, we add an additional group by to get the counts per pid
// func (repository *StatsRepository) GetUniquesByPidForInterval(metricName string, repoId string, interval func(*gorm.DB) *gorm.DB) []PidStat {
// 	var result []PidStat

// 	repository.db.Model(&event.Event{}).
// 		Select("name as metric, pid, count(distinct session_id) as count").
// 		Scopes(NameAndRepoId(metricName, repoId), interval).
// 		Group("name, pid").
// 		Limit(10).
// 		Scan(&result)
// 	return result
// }

// func (repository *StatsRepository) GetUniquesByPidInToday(metricName string, repo_id string) []PidStat {
// 	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIsToday)
// }

// func (repository *StatsRepository) GetUniquesByPidInLast7Days(metricName string, repo_id string) []PidStat {
// 	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIn7Days)
// }

// func (repository *StatsRepository) GetUniquesByPidInLast30Days(metricName string, repo_id string) []PidStat {
// 	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIn30Days)
// }

// Following are scopes for the event model that can be used
// to build up queries dynamically.

func NameAndRepoId(name string, repo_id string) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("name = ?", name).Where("repo_id = ?", repo_id)
	}
}

func RepoId(repo_id string) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("repo_id = ?", repo_id)
	}
}

func PID(pid string) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("pid = ?", pid)
	}
}

// func TimestampIsToday(db *gorm.DB) *gorm.DB {
// 	return db.Where("timestamp > today()")
// }

// func TimestampIn7Days(relativeDate time.Time) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		return db.Where("timestamp > subtractDays(?, 7)", relativeDate)
// 	}
// }

// func TimestampIn30Days(relativeDate time.Time) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		return db.Where("timestamp > subtractDays(?, 30)", relativeDate)
// 	}
// }

func SelectDateByDay(db *gorm.DB) *gorm.DB {
	return db.Select("toStartOfDay(interval_alias) as date")
}

func TimestampCustom(start_date time.Time, end_date time.Time) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("timestamp > ?", start_date).Where("timestamp < ?", end_date)
	}
}

func Paginate(page int, page_size int) func(db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
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
