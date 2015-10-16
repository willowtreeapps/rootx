package rootx

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Rootx defines an interface that needs to be implemented to take advantage
// of the Rootx helper methods.
//
// With the exception of SQL, which is used to get SQL queries by key, these are
// provided by github.com/jmoiron/sqlx's DB and TX.
type Rootx interface {
	SQL(key string) (query string)

	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	Select(instances interface{}, query string, args ...interface{}) error
	Get(instance interface{}, query string, args ...interface{}) error
}

// Error defines an error that can occur in Rootx helper methods.
//
// This is useful to inspect the specific key an error occurred on.
type Error struct {
	Key      string
	Where    string
	SQLError error
}

func (e *Error) Error() string {
	return fmt.Sprintf("Rootx: '%s'-'%s' : %v",
		e.Key, e.Where, e.SQLError.Error())
}

// Exists checks to see if the item exists by a COUNT query
func Exists(db Rootx, sqlKey string, args ...interface{}) (bool, error) {
	var count int
	if err := db.Get(&count, db.SQL(sqlKey), args...); err != nil {
		return false, &Error{sqlKey, "Exists", err}
	}
	return count > 0, nil
}

// SelectAll selects a slice of structs
func SelectAll(db Rootx, sqlKey string, instances interface{}, args ...interface{}) error {
	err := db.Select(instances, db.SQL(sqlKey), args...)
	if err != nil {
		return &Error{sqlKey, "SelectAll", err}
	}
	return nil
}

// SelectOne selects a single struct
func SelectOne(db Rootx, sqlKey string, instance interface{}, args ...interface{}) error {
	rows, err := db.Queryx(db.SQL(sqlKey), args...)
	if err != nil {
		return &Error{sqlKey, "SelectOne-Queryx", err}
	}
	if err = ScanOne(instance, rows); err != nil {
		return &Error{sqlKey, "SelectOne-ScanOne", err}
	}
	return nil
}

// ScanOne returns the instance, if any, returned from sql query
func ScanOne(instance interface{}, rows *sqlx.Rows) error {
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(instance); err != nil {
			return err
		}
	}
	return nil
}

// InsertPsql inserts a single instance
func InsertPsql(db Rootx, sqlKey string, args ...interface{}) (int64, error) {
	row := db.QueryRow(db.SQL(sqlKey), args...)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, &Error{sqlKey, "Insert", err}
	}
	return id, nil
}

// Insert inserts a single instance
func Insert(db Rootx, sqlKey string, args ...interface{}) (int64, error) {
	result, err := db.Exec(db.SQL(sqlKey), args...)
	if err != nil {
		return 0, &Error{sqlKey, "Insert", err}
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, &Error{sqlKey, "Insert-LastInsertId", err}
	}
	return id, nil
}

// DeleteOne deletes a single instance
func DeleteOne(db Rootx, sqlKey string, args ...interface{}) error {
	return updateOne(db, sqlKey, "DeleteOne", args...)
}

// UpdateOne updates a single instance
func UpdateOne(db Rootx, sqlKey string, args ...interface{}) error {
	return updateOne(db, sqlKey, "UpdateOne", args...)
}

// Exec executes a statement
func Exec(db Rootx, sqlKey string, args ...interface{}) error {
	if _, err := db.Exec(db.SQL(sqlKey), args...); err != nil {
		return &Error{sqlKey, "Exec", err}
	}
	return nil
}

func updateOne(db Rootx, sqlKey, where string, args ...interface{}) error {
	res, err := db.Exec(db.SQL(sqlKey), args...)
	if err != nil {
		return &Error{sqlKey, where, err}
	}
	num, err := res.RowsAffected()
	if err != nil {
		return &Error{sqlKey, where + "-RowsAffected", err}
	}
	if num != 1 {
		return &Error{sqlKey, where, fmt.Errorf("%d rows affected, expected 1", num)}
	}
	return nil
}
