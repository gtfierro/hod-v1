package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	logrus "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type VersionManager struct {
	db             *sql.DB
	select_version *sql.Stmt
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

	prepared, err := db.Prepare("select version, name from versions where name = ? order by version limit ?;")
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

func (vm *VersionManager) Graphs() ([]Version, error) {
	rows, err := vm.get_versions.Query()
	defer rows.Close()
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
