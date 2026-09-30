// Package migrations creates the tables the surf lesson API needs.
package migrations

import "gofr.dev/pkg/gofr/migration"

const createInstructors = `CREATE TABLE IF NOT EXISTS instructors (
    id        INTEGER PRIMARY KEY,
    name      TEXT NOT NULL,
    specialty TEXT NOT NULL
);`

const seedInstructors = `INSERT INTO instructors (id, name, specialty) VALUES
    (1, 'Arun', 'Beginners'),
    (2, 'Meera', 'Intermediate'),
    (3, 'Rahul', 'Advanced');`

const createBookings = `CREATE TABLE IF NOT EXISTS bookings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    guest_name    TEXT NOT NULL,
    instructor_id INTEGER NOT NULL REFERENCES instructors(id),
    lesson_date   TEXT NOT NULL,
    slot          TEXT NOT NULL,
    UNIQUE (instructor_id, lesson_date, slot)
);`

// All returns every migration keyed by its timestamp. GoFr runs them in order
// on startup and records which ones have already been applied.
func All() map[int64]migration.Migrate {
	return map[int64]migration.Migrate{
		20260930180000: {
			UP: func(d migration.Datasource) error {
				for _, q := range []string{createInstructors, seedInstructors, createBookings} {
					if _, err := d.SQL.Exec(q); err != nil {
						return err
					}
				}

				return nil
			},
		},
	}
}
