package database

import (
	"context"
	"database/sql"
	"fmt"
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
		`create table if not exists modules (
			id integer primary key,
			name text,

			course_id integer references courses(id)
		)`,
		`create table if not exists components (
			id integer primary key,
			name text,

			module_id integer references modules(id)
		)`,
		`create table if not exists questions (
			id integer,
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

func (db *DB) CreateSubject(subjectID int, subjectName string) error {
	query := `insert or ignore into subjects (id, name) values (?, ?)`
	if _, err := db.Exec(query, subjectID, subjectName); err != nil {
		return fmt.Errorf("CreateSubject: %s", err)
	}

	return nil
}

func (db *DB) CreateCourse(courseID int, courseName string, subjectID int) error {
	query := `insert or ignore into courses (id, name, subject_id) values (?, ?, ?)`
	if _, err := db.Exec(query, courseID, courseName, subjectID); err != nil {
		return fmt.Errorf("CreateCourse: %s", err)
	}

	return nil
}

func (db *DB) CreateModule(moduleID int, moduleName string, courseID int) error {
	query := `insert or ignore into modules (id, name, course_id) values (?, ?, ?)`
	if _, err := db.Exec(query, moduleID, moduleName, courseID); err != nil {
		return fmt.Errorf("CreateModule: %s", err)
	}

	return nil
}

func (db *DB) CreateQuestion(questionID int, questionContent string, rightAnswer int, subjectID, courseID, moduleID int) error {
	query := `insert or ignore into questions (id, content, right_answer, subject_id, course_id, module_id) values (?, ?, ?, ?, ?, ?)`
	if _, err := db.Exec(query, questionID, questionContent, rightAnswer, subjectID, courseID, moduleID); err != nil {
		return fmt.Errorf("CreateQuestion: %s", err)
	}

	return nil
}

func (db *DB) GetAnswer(questionID, subjectID, courseID, moduleID int) (int, error) {
	query := `select right_answer from questions where id = ? and subject_id = ? and course_id = ? and module_id = ?`

	var answer int
	err := db.QueryRow(query, questionID, subjectID, courseID, moduleID).Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return answer, nil
}
