package stats

import (
	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/datacite/keeshond/internal/app/event"
	"gorm.io/gorm"
)

type StatsRepositoryReader interface {
	// Total event metric counts grouped for a specific PID
	GetTotalForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat
	GetTotalInToday(metricName string, repo_id string, pid string) PidStat
	GetTotalInLast7Days(metricName string, repo_id string, pid string) PidStat
	GetTotalInLast30Days(metricName string, repo_id string, pid string) PidStat

	// Total event metric counts grouped by PID
	GetTotalsByPidForInterval(metricName string, repo_id string, interval func(*gorm.DB) *gorm.DB) []PidStat
	GetTotalsByPidInToday(metricName string, repo_id string) []PidStat
	GetTotalsByPidInLast7Days(metricName string, repo_id string) []PidStat
	GetTotalsByPidInLast30Days(metricName string, repo_id string) []PidStat

	// Unique event metric counts for a specific PID
	GetUniqueForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat
	GetUniqueInToday(metricName string, repo_id string, pid string) PidStat
	GetUniqueInLast7Days(metricName string, repo_id string, pid string) PidStat
	GetUniqueInLast30Days(metricName string, repo_id string, pid string) PidStat

	// Unique event metric counts grouped by PID
	GetUniquesByPidForInterval(metricName string, repo_id string, interval func(*gorm.DB) *gorm.DB) []PidStat
	GetUniquesByPidInToday(metricName string, repo_id string) []PidStat
	GetUniquesByPidInLast7Days(metricName string, repo_id string) []PidStat
	GetUniquesByPidInLast30Days(metricName string, repo_id string) []PidStat
}

type StatsRepository struct {
	db 		*gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{
		db: db,
	}
}

// This query is for getting based on an interval the total unique metric events
// We do an inner CTE to be able to remove duplicates based on the interval of 30 secs
// We then can count the returning CTE table to get our total de duplicated events.
// For COUNTER this is referred to as Double Click filtering.
func (repository *StatsRepository) GetTotalForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat {
	var count int64
	repository.db.
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), PID(pid), interval).
			Group("interval_alias"),
		),
	).Table("interval_calc").Count(&count)

	return PidStat {
		Metric: metricName,
		Pid: pid,
		Count: count,
	}
}

func (repository *StatsRepository) GetTotalInToday(metricName string, repo_id string, pid string) PidStat {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIsToday)
}

func (repository *StatsRepository) GetTotalInLast7Days(metricName string, repo_id string, pid string) PidStat {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn7Days)
}

func (repository *StatsRepository) GetTotalInLast30Days(metricName string, repo_id string, pid string) PidStat {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn30Days)
}

// Similiar to the query for specific pid, builds a CTE that is grouped by the interval of 30 secs to remove duplicates in interval
// But also returns the pid and then we do a group by again to get counts for pids.
func (repository *StatsRepository) GetTotalsByPidForInterval(metricName string, repoId string, interval func(*gorm.DB) *gorm.DB) []PidStat {
	var result []PidStat

	repository.db.Debug().
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("name, pid, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), interval).
			Group("name, pid, interval_alias"),
		),
	).Table("interval_calc").Select("name as metric, pid, count(*) as count").Group("name, pid").Limit(10).Scan(&result)

	return result
}

func (repository *StatsRepository) GetTotalsByPidInToday(metricName string, repo_id string) []PidStat {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIsToday)
}

func (repository *StatsRepository) GetTotalsByPidInLast7Days(metricName string, repo_id string) []PidStat {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn7Days)
}

func (repository *StatsRepository) GetTotalsByPidInLast30Days(metricName string, repo_id string) []PidStat {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn30Days)
}

// To get the unique values, this can be achieved with a simple distinct query across events for specific pids/repos
// This doesn't need the deduping logic because sessions are always same within 1 hour, therefore any duplicated events
// will not appear in the result count.
// For COUNTER this is referred to as Counting Unique Datasets
func (repository *StatsRepository) GetUniqueForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) PidStat {
	var count int64
	repository.db.Model(&event.Event{}).
		Distinct("session_id").
		Scopes(NameAndRepoId(metricName, repoId), PID(pid), interval).
		Count(&count)

	return PidStat {
		Metric: metricName,
		Pid: pid,
		Count: count,
	}
}

func (repository *StatsRepository) GetUniqueInToday(metricName string, repo_id string, pid string) PidStat {
	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIsToday)
}

func (repository *StatsRepository) GetUniqueInLast7Days(metricName string, repo_id string, pid string) PidStat {
	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIn7Days)
}

func (repository *StatsRepository) GetUniqueInLast30Days(metricName string, repo_id string, pid string) PidStat {
	return repository.GetUniqueForInterval(metricName, repo_id, pid, TimestampIn30Days)
}

// Similiar to the unique values by PID, we can do the same as this except, we add an additional group by to get the counts per pid
func (repository *StatsRepository) GetUniquesByPidForInterval(metricName string, repoId string, interval func(*gorm.DB) *gorm.DB) []PidStat {
	var result []PidStat

	repository.db.Model(&event.Event{}).
		Select("name as metric, pid, count(distinct session_id) as count").
		Scopes(NameAndRepoId(metricName, repoId), interval).
		Group("name, pid").
		Limit(10).
		Scan(&result)
	return result
}

func (repository *StatsRepository) GetUniquesByPidInToday(metricName string, repo_id string) []PidStat {
	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIsToday)
}

func (repository *StatsRepository) GetUniquesByPidInLast7Days(metricName string, repo_id string) []PidStat {
	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIn7Days)
}

func (repository *StatsRepository) GetUniquesByPidInLast30Days(metricName string, repo_id string) []PidStat {
	return repository.GetUniquesByPidForInterval(metricName, repo_id, TimestampIn30Days)
}

// Following are scopes for the event model that can be used
// to build up queries dynamically.

func NameAndRepoId(name string, repo_id string) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("name = ?", name).Where("repo_id = ?", repo_id)
	}
}

func PID(pid string) func (db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
	  return db.Where("pid = ?", pid)
	}
}

func TimestampIsToday(db *gorm.DB) *gorm.DB {
	return db.Where("timestamp > today()")
}

func TimestampIn7Days(db *gorm.DB) *gorm.DB {
	return db.Where("timestamp > subtractDays(now(), 7)")
}

func TimestampIn30Days(db *gorm.DB) *gorm.DB {
	return db.Where("timestamp > subtractDays(now(), 30)")
}