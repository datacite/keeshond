package stats

import (
	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/datacite/keeshond/internal/app/event"
	"gorm.io/gorm"
)

type StatsRepositoryReader interface {
	// Total event metric counts grouped for a specific PID
	GetTotalForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) int64
	GetTotalInToday(metricName string, repo_id string, pid string) int64
	GetTotalInLast7Days(metricName string, repo_id string, pid string) int64
	GetTotalInLast30Days(metricName string, repo_id string, pid string) int64

	// Total event metric counts grouped by PID
	GetTotalsByPidForInterval(metricName string, repo_id string, interval func(*gorm.DB) *gorm.DB) []struct{Pid string; Count int64}
	GetTotalsByPidInToday(metricName string, repo_id string) []struct{Pid string; Count int64}
	GetTotalsByPidInLast7Days(metricName string, repo_id string) []struct{Pid string; Count int64}
	GetTotalsByPidInLast30Days(metricName string, repo_id string) []struct{Pid string; Count int64}
	// GetUniqueInToday(metricName string, repo_id string, pid string) int64
	// GetUniqueInLast7Days(metricName string, repo_id string, pid string) int64
	// GetUniqueInLast30Days(metricName string, repo_id string, pid string) int64
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
func (repository *StatsRepository) GetTotalForInterval(metricName string, repoId string, pid string, interval func(*gorm.DB) *gorm.DB) int64 {
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
	return count
}

func (repository *StatsRepository) GetTotalInToday(metricName string, repo_id string, pid string) int64 {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIsToday)
}

func (repository *StatsRepository) GetTotalInLast7Days(metricName string, repo_id string, pid string) int64 {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn7Days)
}

func (repository *StatsRepository) GetTotalInLast30Days(metricName string, repo_id string, pid string) int64 {
	return repository.GetTotalForInterval(metricName, repo_id, pid, TimestampIn30Days)
}

// Similiar to the query for specific pid, builds a CTE that is grouped by the interval of 30 secs to remove duplicates in interval
// But also returns the pid and then we do a group by again to get counts for pids.
func (repository *StatsRepository) GetTotalsByPidForInterval(metricName string, repoId string, interval func(*gorm.DB) *gorm.DB) []struct{Pid string; Count int64} {
	var result []struct {
		Pid string
		Count int64
	}
	repository.db.
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("pid, toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), interval).
			Group("pid, interval_alias"),
		),
	).Table("interval_calc").Select("pid, count(*) as count").Group("pid").Limit(10).Scan(&result)
	return result
}

func (repository *StatsRepository) GetTotalsByPidInToday(metricName string, repo_id string) []struct{Pid string; Count int64} {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIsToday)
}

func (repository *StatsRepository) GetTotalsByPidInLast7Days(metricName string, repo_id string) []struct{Pid string; Count int64} {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn7Days)
}

func (repository *StatsRepository) GetTotalsByPidInLast30Days(metricName string, repo_id string) []struct{Pid string; Count int64} {
	return repository.GetTotalsByPidForInterval(metricName, repo_id, TimestampIn30Days)
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