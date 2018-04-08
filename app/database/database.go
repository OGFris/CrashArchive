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
