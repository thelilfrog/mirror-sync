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

	for _, repo := range pr.Repositories {
		if err := r.createRepository(tx, projectUUID, repo); err != nil {
			return err
		}
	}

	return nil
}

func (r Repository) createRepository(tx *sql.Tx, projectUuid string, repo project.Repository) error {
	repoUUID := uuid.NewString()

	stmt, err := tx.Prepare("INSERT INTO Repositories (uuid, name, source, destination, schedule, project) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to create statement: %s", err)
	}

	if _, err := stmt.Exec(repoUUID, repo.Name, repo.Source, repo.Destination, repo.Schedule, projectUuid); err != nil {
		return fmt.Errorf("failed to execute sql query: %s", err)
	}

	for ref, auth := range repo.Authentications {
		if len(auth.Token) > 0 {
			stmt, err := tx.Prepare("INSERT INTO Authentication (repository, ref, token) VALUES (?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to create statement: %s", err)
			}

			if _, err := stmt.Exec(repoUUID, ref, auth.Token); err != nil {
				return fmt.Errorf("failed to execute sql query: %s", err)
			}
		} else if auth.Basic != nil {
			stmt, err := tx.Prepare("INSERT INTO Authentication (repository, ref, username, password) VALUES (?, ?, ?, ?)")
			if err != nil {
				return fmt.Errorf("failed to create statement: %s", err)
			}

			if _, err := stmt.Exec(repoUUID, ref, auth.Basic.Username, auth.Basic.Password); err != nil {
				return fmt.Errorf("failed to execute sql query: %s", err)
			}
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

	// this loop does NOT remove orphan
	for _, repo := range pr.Repositories {
		// checks if the repo exists
		exists, err := r.RepositoryExistsByName(repo.Name)
		if err != nil {
			return fmt.Errorf("failed to fetch uuid from the database: %w", err)
		}

		if exists {
			// if it exists, just update it
			if err := r.updateRepository(tx, repo); err != nil {
				return err
			}
		} else {
			// if not, create a new uuid and create the entry
			if err := r.createRepository(tx, projectUUID, repo); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) updateRepository(tx *sql.Tx, repo project.Repository) error {
	uuid, err := r.RepositoryUUID(repo.Name)
	if err != nil {
		return fmt.Errorf("failed to get uuid from database: %w", err)
	}

	stmt, err := tx.Prepare("UPDATE Repositories SET schedule = ?, source = ?, destination = ? WHERE uuid = ?")
	if err != nil {
		return fmt.Errorf("failed to create statement: %w", err)
	}

	if _, err := stmt.Exec(repo.Schedule, repo.Source, repo.Destination, uuid); err != nil {
		return fmt.Errorf("failed to update repository entry for %s::'%s'", uuid, repo.Name)
	}

	for ref, auth := range repo.Authentications {
		if auth.Basic != nil {
			stmt, err := tx.Prepare("UPDATE Authentication SET username = ?, password = ?, token = null WHERE repository = ? AND ref = ?")
			if err != nil {
				return fmt.Errorf("failed to create statement: %w", err)
			}
			if _, err := stmt.Exec(auth.Basic.Username, auth.Basic.Password, uuid, ref); err != nil {
				return fmt.Errorf("failed to execut sql query: %s", err)
			}
		} else {
			stmt, err := tx.Prepare("UPDATE Authentication SET username = null, password = null, token = ? WHERE repository = ? AND ref = ?")
			if err != nil {
				return fmt.Errorf("failed to create statement: %w", err)
			}
			if _, err := stmt.Exec(auth.Token, uuid, ref); err != nil {
				return fmt.Errorf("failed to execute sql query: %s", err)
			}
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

	stmt, err := r.db.Prepare("SELECT uuid, name, schedule, source, destination FROM Repositories WHERE project = ?")
	if err != nil {
		return nil, fmt.Errorf("invalid syntax: %w", err)
	}

	authStmt, err := r.db.Prepare("SELECT ref, username, password, token FROM Authentication WHERE repository = ?")
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
			var uuid string
			var repo project.Repository
			if err := repoRows.Scan(&uuid, &repo.Name, &repo.Schedule, &repo.Source, &repo.Destination); err != nil {
				repoRows.Close()
				return nil, fmt.Errorf("failed to scan repository entry: %w", err)
			}

			authRows, err := authStmt.Query(uuid)
			if err != nil {
				repoRows.Close()
				return nil, fmt.Errorf("failed to query repositories for the project %s: %w", prUUID, err)
			}

			auth := make(map[string]project.AuthenticationSettings)
			for authRows.Next() {
				var ref string
				var username, password, token *string
				if err := authRows.Scan(&ref, &username, &password, &token); err != nil {
					authRows.Close()
					repoRows.Close()
					return nil, fmt.Errorf("failed to scan authentication entry: %s", err)
				}
				if token != nil {
					auth[ref] = project.AuthenticationSettings{
						Token: *token,
					}
				} else if username != nil {
					auth[ref] = project.AuthenticationSettings{
						Basic: &project.BasicAuthenticationSettings{
							Username: *username,
							Password: *password,
						},
					}
				}
			}

			authRows.Close()
			repo.Authentications = auth
			pr.Repositories = append(pr.Repositories, repo)
		}

		repoRows.Close()
		prs = append(prs, pr)
	}

	return prs, nil
}
