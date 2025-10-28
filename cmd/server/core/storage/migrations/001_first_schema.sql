-- +goose Up
CREATE TABLE Projects (
	uuid TEXT NOT NULL,
	name TEXT NOT NULL
);
CREATE INDEX Projects_uuid_IDX ON Projects (uuid);

CREATE TABLE Repositories (
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	schedule TEXT NOT NULL,
	"source" TEXT NOT NULL,
	destination TEXT NOT NULL,
    project INTEGER NOT NULL
);
CREATE INDEX Repositories_uuid_IDX ON Repositories (uuid);

CREATE TABLE Authentication (
	repository TEXT NOT NULL,
	ref TEXT NOT NULL,
	username TEXT,
	"password" TEXT,
	token TEXT
);

-- +goose Down
DROP TABLE Projects;
DROP TABLE Repositories;