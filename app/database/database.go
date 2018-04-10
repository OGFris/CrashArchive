package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pmmp/CrashArchive/app/crashreport"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	retryTimes = 5
	retryDelay = 5 * time.Second
)

type DB struct {
	*sqlx.DB
}

func New(config *Config) (*DB, error) {
	if config.Username == "" || config.Password == "" {
		return nil, errors.New("Username and password for mysql database not set in config.json")
	}

	for i := 1; i <= retryTimes; i++ {
		log.Printf("Connecting to database %d/%d", i, retryTimes)
		db, err := sqlx.Connect("mysql", DSN(config))
		if err == nil {
			log.Printf("Connected to database")
			return &DB{db}, nil
		}
		log.Println(err)
		if i != retryTimes {
			time.Sleep(retryDelay)
		}
	}

	return nil, errors.New("failed to connect")
}

var queryInsertReport = `INSERT INTO crash_reports
		(plugin, version, build, file, message, line, type, os, submitDate, reportDate, duplicate)
	VALUES
		(:plugin, :version, :build, :file, :message, :line, :type, :os, :submitDate, :reportDate, :duplicate)`

func (db *DB) InsertReport(report *crashreport.CrashReport) (int64, error) {
	res, err := db.NamedExec(queryInsertReport, &crashreport.Report{
		Plugin:     report.CausingPlugin,
		Version:    report.Version.Get(true),
		Build:      report.Version.Build,
		File:       report.Error.File,
		Message:    report.Error.Message,
		Line:       report.Error.Line,
		Type:       report.Error.Type,
		OS:         report.Data.General.OS,
		SubmitDate: time.Now().Unix(),
		ReportDate: report.ReportDate.Unix(),
		Duplicate:  report.Duplicate,
	})

	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("failed to get last insert ID: %d", id)
	}

	return id, nil
}

func (db *DB) CheckDuplicate(report *crashreport.CrashReport) (int, error) {
	queryDupe := "SELECT COUNT(id) FROM (SELECT id, message, file, line FROM crash_reports ORDER BY id DESC LIMIT 5000)sub WHERE message = ? AND file = ? and line = ?;"

	var dupes int
	err := db.Get(&dupes, queryDupe, report.Error.Message, report.Error.File, report.Error.Line)
	if err != nil {
		return 0, err
	}

	return dupes, nil
}

func (db *DB) GetFilteredReports(pageID, pageSize int, filter string, params ...interface{}) (int, int, []crashreport.Report, error) {
	var reports []crashreport.Report
	var total int

	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM crash_reports %s", filter)
	if err := db.Get(&total, queryCount, params...); err != nil {
		log.Println(queryCount)
		return 0, 0, reports, err
	}

	if (pageID-1)*pageSize > total {
		return 0, 0, reports, errors.New("one page too many")
	}

	rangeStart := (pageID - 1) * pageSize
	querySelect := fmt.Sprintf("SELECT id, version, message FROM crash_reports %s ORDER BY id DESC LIMIT %d, %d", filter, rangeStart, pageSize)
	if err := db.Select(&reports, querySelect, params...); err != nil {
		log.Println(querySelect)
		return 0, 0, reports, err
	}

	return total, rangeStart, reports, nil
}
