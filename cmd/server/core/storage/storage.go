package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"mirror-sync/pkg/project"

	_ "github.com/glebarez/go-sqlite"
	"github.com/google/uuid"
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
	exists, err := r.ProjectExistsByName(pr.Name)
	if err != nil {
		return err
	}

	if exists {
		return r.Update(pr)
	}

	return r.Create(pr)
}

func (r *Repository) ProjectExistsByUUID(uuid string) (bool, error) {
	row := r.db.QueryRow("SELECT uuid FROM Projects WHERE uuid = ?", uuid)
	if row.Err() != nil {
		return false, fmt.Errorf("failed to get row from database: %w", row.Err())
	}

	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to scan row: %w", err)
	}

	return true, nil
}

func (r *Repository) ProjectExistsByName(name string) (bool, error) {
	row := r.db.QueryRow("SELECT uuid FROM Projects WHERE name = ?", name)
	if row.Err() != nil {
		return false, fmt.Errorf("failed to get row from database: %w", row.Err())
	}

	var uuid string
	if err := row.Scan(&uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to scan row: %w", err)
	}

	return true, nil
}

func (r *Repository) RepositoryExistsByName(name string) (bool, error) {
	row := r.db.QueryRow("SELECT uuid FROM Repositories WHERE name = ?", name)
	if row.Err() != nil {
		return false, fmt.Errorf("failed to get row from database: %w", row.Err())
	}

	var uuid string
	if err := row.Scan(&uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to scan row: %w", err)
	}

	return true, nil
}

func (r *Repository) Create(pr project.Project) error {
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

	// Create Project entry
	projectUUID := uuid.NewString()

	stmt, err := tx.Prepare("INSERT INTO Projects (uuid, name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to create statement: %s", err)
	}
	if _, err := stmt.Exec(projectUUID, pr.Name); err != nil {
		return fmt.Errorf("failed to execute sql query: %s", err)
	}

	// Create repositories entries
	stmt, err = tx.Prepare("INSERT INTO Repositories (uuid, name, source, destination, schedule, project) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to create statement: %s", err)
	}

	for _, repo := range pr.Repositories {
		repoUUID := uuid.NewString()

		if _, err := stmt.Exec(repoUUID, repo.Name, repo.Source, repo.Destination, repo.Schedule, projectUUID); err != nil {
			return fmt.Errorf("failed to execute sql query: %s", err)
		}
	}

	return nil
}

func (r *Repository) ProjectUUID(name string) (string, error) {
	row := r.db.QueryRow("SELECT uuid FROM Projects WHERE name = ?", name)
	if row.Err() != nil {
		return "", fmt.Errorf("failed to get row from database: %w", row.Err())
	}

	var uuid string
	if err := row.Scan(&uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	return uuid, nil
}

func (r *Repository) RepositoryUUID(name string) (string, error) {
	row := r.db.QueryRow("SELECT uuid FROM Repositories WHERE name = ?", name)
	if row.Err() != nil {
		return "", fmt.Errorf("failed to get row from database: %w", row.Err())
	}

	var uuid string
	if err := row.Scan(&uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	return uuid, nil
}

func (r *Repository) Update(pr project.Project) error {
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

	projectUUID, err := r.ProjectUUID(pr.Name)
	if err != nil {
		return fmt.Errorf("failed to get project uuid: %w", err)
	}

	stmt, err := tx.Prepare("UPDATE Repositories SET schedule = ?, source = ?, destination = ? WHERE uuid = ?")
	if err != nil {
		return fmt.Errorf("failed to create statement: %w", err)
	}

	// this loop does NOT remove orphan
	for _, repo := range pr.Repositories {
		// checks if the repo exists
		exists, err := r.RepositoryExistsByName(repo.Name)
		if err != nil {
			return fmt.Errorf("failed to fetch uuid from the database: %w", err)
		}

		if exists {
			// if it exists, just update it
			uuid, err := r.RepositoryUUID(repo.Name)
			if err != nil {
				return fmt.Errorf("failed to get uuid from database: %w", err)
			}

			if _, err := stmt.Exec(repo.Schedule, repo.Source, repo.Destination, uuid); err != nil {
				return fmt.Errorf("failed to update repository entry for %s::'%s'", uuid, repo.Name)
			}
		} else {
			// if not, create a new uuid and create the entry
			repoUUID := uuid.NewString()

			if _, err := stmt.Exec(repoUUID, repo.Name, repo.Source, repo.Destination, repo.Schedule, projectUUID); err != nil {
				return fmt.Errorf("failed to execute sql query: %s", err)
			}
		}
		if _, err := stmt.Exec(repo.Schedule, repo.Source, repo.Destination, repo.Name); err != nil {
			return fmt.Errorf("failed to update repository entry for '%s'", repo.Name)
		}
	}

	return nil
}

func (r *Repository) List() ([]project.Project, error) {
	var prs []project.Project

	rows, err := r.db.Query("SELECT uuid, name FROM Projects")
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of projects: %w", err)
	}
	defer rows.Close()

	stmt, err := r.db.Prepare("SELECT name, schedule, source, destination FROM Repositories WHERE project = ?")
	if err != nil {
		return nil, fmt.Errorf("invalid syntax: %w", err)
	}
	for rows.Next() {
		var pr project.Project
		var prUUID string
		if err := rows.Scan(&prUUID, &pr.Name); err != nil {
			return nil, fmt.Errorf("failed to scan project name: %w", err)
		}

		repoRows, err := stmt.Query(prUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to query repositories for the project %s: %w", prUUID, err)
		}

		for repoRows.Next() {
			var repo project.Repository
			if err := repoRows.Scan(&repo.Name, &repo.Schedule, &repo.Source, &repo.Destination); err != nil {
				repoRows.Close()
				return nil, fmt.Errorf("failed to scan repository entry: %w", err)
			}
			pr.Repositories = append(pr.Repositories, repo)
		}

		repoRows.Close()
		prs = append(prs, pr)
	}

	return prs, nil
}
