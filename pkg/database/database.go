package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = db.Exec(`pragma foreign_keys = ON;`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	stmts := []string{
		`create table if not exists subjects (id integer primary key, name text)`,
		`create table if not exists courses (
			id integer primary key,
			name text,

			subject_id integer references subjects(id)
		)`,
		`create table if not exists modules (id integer primary key, name text)`,
		`create table if not exists components (
			id integer primary key,
			name text,

			module_id integer references modules(id)
		)`,
		`create table if not exists questions (
			id integer primary key,
			content text,
			right_answer integer,

			subject_id integer references subjects(id),
			course_id integer references courses(id),
			module_id integer references modules(id)
		)`,
	}

	tx, err := db.Begin()
	if err != nil {
		db.Close()
		return nil, err
	}

	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			tx.Rollback()
			db.Close()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{db}, nil
}
