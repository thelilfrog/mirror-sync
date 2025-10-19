package storage

import (
	"database/sql"
	"fmt"
	"mirror-sync/pkg/project"

	_ "github.com/glebarez/go-sqlite"
)

type (
	Repository struct {
		db *sql.DB
	}
)

func OpenDB(path string) (*Repository, error) {
	// connect
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err)
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) Save(pr project.Project) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to create transaction: %s", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	stmt, err := tx.Prepare("INSERT INTO Projects (name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to create statement: %s", err)
	}
	if _, err := stmt.Exec(pr.Name); err != nil {
		return fmt.Errorf("failed to execute sql query: %s", err)
	}

	rows, err := tx.Query("SELECT id FROM Projects WHERE name = ?", pr.Name)
	if err != nil {
		return fmt.Errorf("failed to query project id: %s", err)
	}
	defer rows.Close()

	var id int
	rows.Next()
	if err := rows.Scan(&id); err != nil {
		return fmt.Errorf("failed to query project id: %s", err)
	}

	for _, repo := range pr.Repositories {
		stmt, err := tx.Prepare("INSERT INTO Repositories (name, source, destination, schedule, project) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to create statement: %s", err)
		}
		if _, err := stmt.Exec(repo.Name, repo.Source, repo.Destination, repo.Schedule, id); err != nil {
			return fmt.Errorf("failed to execute sql query: %s", err)
		}
	}

	return nil
}
