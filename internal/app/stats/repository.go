package stats

import (
	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/datacite/keeshond/internal/app/event"
	"gorm.io/gorm"
)

type StatsRepositoryReader interface {
	GetTotalInToday(metricName string, repo_id string, pid string) int64
	GetTotalInLast7Days(metricName string, repo_id string, pid string) int64
	GetTotalInLast30Days(metricName string, repo_id string, pid string) int64
	// GetUniqueInToday(metricName string, repo_id string, pid string) int64
	// GetUniqueInLast7Days(metricName string, repo_id string, pid string) int64
	// GetUniqueInLast30Days(metricName string, repo_id string, pid string) int64
}

type StatsRepository struct {
	db 		*gorm.DB
}

func NewRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{
		db: db,
	}
}

func (repository *StatsRepository) GetTotalInToday(metricName string, repoId string, pid string) int64 {
	var count int64
	repository.db.
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), PID(pid), TimestampIsToday).
			Group("interval_alias"),
		),
	).Table("interval_calc").Count(&count)
	return count
}

func (repository *StatsRepository) GetTotalInLast7Days(metricName string, repoId string, pid string) int64 {
	var count int64
	repository.db.
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), PID(pid), TimestampIn7Days).
			Group("interval_alias"),
		),
	).Table("interval_calc").Count(&count)
	return count
}

func (repository *StatsRepository) GetTotalInLast30Days(metricName string, repoId string, pid string) int64 {
	var count int64
	repository.db.
	Clauses(
		exclause.NewWith(
			"interval_calc", repository.db.Model(&event.Event{}).
			Select("toStartOfInterval(timestamp, INTERVAL 30 second) as interval_alias").
			Scopes(NameAndRepoId(metricName, repoId), PID(pid), TimestampIn30Days).
			Group("interval_alias"),
		),
	).Table("interval_calc").Count(&count)
	return count
}


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