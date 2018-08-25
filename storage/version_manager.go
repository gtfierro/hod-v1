package storage

import (
	"database/sql"
	"github.com/pkg/errors"
	// import sqlite driver
	_ "github.com/mattn/go-sqlite3"
	logrus "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type VersionManager struct {
	db             *sql.DB
	select_version *sql.Stmt
	name_query     *sql.Stmt
	select_at      *sql.Stmt
	select_before  *sql.Stmt
	select_after   *sql.Stmt
	get_versions   *sql.Stmt
}

func CreateVersionManager(dir string) (*VersionManager, error) {
	db, err := sql.Open("sqlite3", filepath.Join(dir, "version.sqlite3"))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	sqlStmt := `
    create table if not exists versions (version integer not null primary key, name text);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	prepared, err := db.Prepare("select version, name from versions where name = ? order by version desc limit ?;")
	if err != nil {
		return nil, err
	}

	name_query, err := db.Prepare("select distinct(name) from versions;")
	if err != nil {
		return nil, err
	}

	prepared_at, err := db.Prepare("select version, name from versions where name = ? and version < ? order by version desc limit 1;")
	if err != nil {
		return nil, err
	}

	prepared_before, err := db.Prepare("select version, name from versions where name = ? and version < ? order by version desc limit 1,?;")
	if err != nil {
		return nil, err
	}

	prepared_after, err := db.Prepare("select version, name from versions where name = ? and version > ? order by version desc limit ?;")
	if err != nil {
		return nil, err
	}

	prepared_listall, err := db.Prepare("select max(version), name from versions group by name;")
	if err != nil {
		return nil, err
	}

	return &VersionManager{
		db:             db,
		select_version: prepared,
		name_query:     name_query,
		select_at:      prepared_at,
		select_before:  prepared_before,
		select_after:   prepared_after,
		get_versions:   prepared_listall,
	}, nil
}

func (vm *VersionManager) GetLatestVersion(name string) (latest Version, err error) {
	rows, err := vm.select_version.Query(name, "1")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&latest.Timestamp, &latest.Name)
		logrus.Infof("Got version in latestversion: %+v", latest)
		return
	}
	return
}

func (vm *VersionManager) NewVersion(name string) (newest Version, err error) {
	//TODO: transaction?
	stmt, err := vm.db.Prepare("insert into versions(version, name) values(?, ?)")
	if err != nil {
		return
	}

	version := uint64(time.Now().UnixNano())
	if _, err = stmt.Exec(version, name); err != nil {
		return
	}
	return Version{version, name}, nil
}

func (vm *VersionManager) ListVersions(name string) (versions []Version, err error) {
	stmt, err := vm.db.Prepare("select distinct version from versions where name = ?")
	if err != nil {
		return
	}

	rows, err := stmt.Query(name)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		var ver = Version{Name: name}
		err = rows.Scan(&ver.Timestamp)
		if err != nil {
			return
		}
		versions = append(versions, ver)

	}
	return
}

func (vm *VersionManager) AddVersion(v Version) (err error) {
	//TODO: transaction?
	stmt, err := vm.db.Prepare("insert into versions(version, name) values(?, ?)")
	if err != nil {
		return
	}

	if _, err = stmt.Exec(v.Timestamp, v.Name); err != nil {
		return
	}
	return
}

// returns the names of all graphs that have at least 1 version
func (vm *VersionManager) Names() ([]string, error) {
	rows, err := vm.name_query.Query()
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var names []string
	for rows.Next() {
		var n string
		if err = rows.Scan(&n); err != nil {
			return names, err
		}
		names = append(names, n)
	}
	return names, nil

}

func (vm *VersionManager) Graphs() ([]Version, error) {
	rows, err := vm.get_versions.Query()
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var versions []Version
	for rows.Next() {
		var v Version
		if err = rows.Scan(&v.Timestamp, &v.Name); err != nil {
			return versions, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (vm *VersionManager) GetVersionAt(name string, t time.Time) (version Version, err error) {
	rows, err := vm.select_at.Query(name, t.UnixNano())
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&version.Timestamp, &version.Name)
		return
	}
	return
}

func (vm *VersionManager) GetVersionsBefore(name string, t time.Time, limit int) (versions []Version, err error) {
	rows, err := vm.select_before.Query(name, t.UnixNano(), limit)
	if err != nil {
		err = errors.Wrap(err, "no query")
		return
	}
	defer rows.Close()
	// current version
	for rows.Next() {
		var v Version
		if err = rows.Scan(&v.Timestamp, &v.Name); err != nil {
			return
		}
		versions = append(versions, v)
	}
	return
}

func (vm *VersionManager) GetVersionsAfter(name string, t time.Time, limit int) (versions []Version, err error) {
	rows, err := vm.select_after.Query(name, t.UnixNano(), limit)
	if err != nil {
		return
	}
	defer rows.Close()
	// current version
	for rows.Next() {
		var v Version
		if err = rows.Scan(&v.Timestamp, &v.Name); err != nil {
			return
		}
		versions = append(versions, v)
	}
	return
}
