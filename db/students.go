package db

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) AddStudent(ctx context.Context, student *Student) error {
	query := `
        INSERT INTO students (uid, status, name, start_date, end_date, schedule, excluded_days, break_time)
        VALUES (?,?,?,?,?,?,?,?)
    `
	res, err := s.db.ExecContext(
		ctx, query,
		student.UID, student.Status, student.Name, student.StartDate, student.EndDate,
		student.Schedule, student.ExcludedDays, student.BreakTime,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	student.ID = int(id)
	return nil
}

func (s *Store) ListStudents(ctx context.Context) ([]*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, created_at 
		FROM students
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*Student, 0)
	for rows.Next() {
		st := &Student{}
		err := rows.Scan(
			&st.ID, &st.UID, &st.Name, &st.Status, &st.StartDate,
			&st.EndDate, &st.Schedule, &st.DoneSeconds,
			&st.ExcludedDays, &st.BreakTime, &st.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, st)
	}

	return students, rows.Err()
}

func (s *Store) FindStudentByID(ctx context.Context, id int) (*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, created_at
		FROM students
		WHERE id = ? LIMIT 1
	`

	var st Student
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&st.ID, &st.UID, &st.Name, &st.Status, &st.StartDate,
		&st.EndDate, &st.Schedule, &st.DoneSeconds,
		&st.ExcludedDays, &st.BreakTime, &st.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no student found based on id: %d", id)
		}
		return nil, err
	}

	return &st, nil
}

func (s *Store) FindStudentByUID(ctx context.Context, uid string) (*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, created_at
		FROM students
		WHERE uid = ? LIMIT 1
	`

	var st Student
	err := s.db.QueryRowContext(ctx, query, uid).Scan(
		&st.ID, &st.UID, &st.Name, &st.Status, &st.StartDate,
		&st.EndDate, &st.Schedule, &st.DoneSeconds,
		&st.ExcludedDays, &st.BreakTime, &st.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no student found based on rfid uid: %s", uid)
		}
		return nil, err
	}

	return &st, nil
}

func (s *Store) UpdateStudent(ctx context.Context, id int, student Student) error {
	query := `
		UPDATE students
		SET uid = ?,
			status = ?,
			name = ?,
			start_date = ?,
			end_date = ?,
			schedule = ?,
			excluded_days = ?,
			break_time = ?
		WHERE id = ?
    `
	_, err := s.db.ExecContext(
		ctx, query,
		student.UID, student.Status, student.Name, student.StartDate, student.EndDate,
		student.Schedule, student.ExcludedDays, student.BreakTime, id,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteStudentByID(ctx context.Context, id int) error {
	query := `
		DELETE FROM students
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IncrementDoneSeconds(ctx context.Context, seconds int) error {
	query := `
		UPDATE students 
		SET done_seconds = done_seconds + ? 
		WHERE status = 'IN'
	`
	_, err := s.db.ExecContext(ctx, query, seconds)
	return err
}
