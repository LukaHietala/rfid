package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"time"
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

	remaining, err := calculateRemainingSeconds(*student)
	if err != nil {
		return err
	}
	student.Remaining = remaining

	return nil
}

func (s *Store) ListStudents(ctx context.Context, archived bool) ([]*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, is_archived, created_at 
		FROM students
	`

	if !archived {
		query += `WHERE is_archived = FALSE`
	} else {
		query += `WHERE is_archived = TRUE`
	}

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
			&st.ExcludedDays, &st.BreakTime, &st.IsArchived, &st.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		students = append(students, st)
	}

	for _, s := range students {
		remaining, err := calculateRemainingSeconds(*s)
		if err != nil {
			return nil, err
		}
		s.Remaining = remaining
	}

	return students, rows.Err()
}

func (s *Store) FindStudentByID(ctx context.Context, id int) (*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, is_archived, created_at
		FROM students
		WHERE id = ? LIMIT 1
	`

	var st Student
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&st.ID, &st.UID, &st.Name, &st.Status, &st.StartDate,
		&st.EndDate, &st.Schedule, &st.DoneSeconds,
		&st.ExcludedDays, &st.BreakTime, &st.IsArchived, &st.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no student found based on id: %d", id)
		}
		return nil, err
	}

	remaining, err := calculateRemainingSeconds(st)
	if err != nil {
		return nil, err
	}
	st.Remaining = remaining

	return &st, nil
}

func (s *Store) FindStudentByUID(ctx context.Context, uid string) (*Student, error) {
	query := `
		SELECT id, uid, name, status, start_date, end_date, schedule, done_seconds, excluded_days, break_time, is_archived, created_at
		FROM students
		WHERE uid = ? LIMIT 1
	`

	var st Student
	err := s.db.QueryRowContext(ctx, query, uid).Scan(
		&st.ID, &st.UID, &st.Name, &st.Status, &st.StartDate,
		&st.EndDate, &st.Schedule, &st.DoneSeconds,
		&st.ExcludedDays, &st.BreakTime, &st.IsArchived, &st.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no student found based on rfid uid: %s", uid)
		}
		return nil, err
	}

	remaining, err := calculateRemainingSeconds(st)
	if err != nil {
		return nil, err
	}
	st.Remaining = remaining

	return &st, nil
}

func (s *Store) UpdateStudent(ctx context.Context, id int, student *Student) error {
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

	remaining, err := calculateRemainingSeconds(*student)
	if err != nil {
		return err
	}
	student.ID = id
	student.Remaining = remaining

	return nil
}

func (s *Store) ArchiveStudentByID(ctx context.Context, id int) error {
	query := `
		UPDATE students
		SET is_archived = TRUE
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UnarchiveStudentByID(ctx context.Context, id int) error {
	query := `
		UPDATE students
		SET is_archived = FALSE
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// Should be avoided
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
		WHERE status = 'IN' AND date(current_timestamp, 'localtime') <= end_date
	`
	_, err := s.db.ExecContext(ctx, query, seconds)
	return err
}

func calculateRemainingSeconds(s Student) (int, error) {
	now := time.Now()
	start, err := time.Parse(time.DateOnly, s.StartDate)
	if err != nil {
		return 0, fmt.Errorf("unable to parse start date: %s: %w", s.StartDate, err)
	}
	end, err := time.Parse(time.DateOnly, s.EndDate)
	if err != nil {
		return 0, fmt.Errorf("unable to parse end date: %s: %w", s.EndDate, err)
	}

	if !end.After(start) {
		return 0, fmt.Errorf("start date must be greater than end date! start: %s end: %s", start, end)
	}

	excludedDays := make([]time.Time, 0, len(s.ExcludedDays))
	for _, v := range s.ExcludedDays {
		day, err := time.Parse(time.DateOnly, v)
		if err != nil {
			log.Println("failed to parse excluded", err)
			continue
		}
		excludedDays = append(excludedDays, day)
	}

	type parsedDay struct {
		duration time.Duration
	}
	parsedSchedule := make(map[int]parsedDay)

	for weekday, day := range s.Schedule {
		if day == nil {
			continue
		}
		dayStarts, err := time.Parse(time.TimeOnly, day.Start)
		if err != nil {
			log.Printf("failed to parse start date for weekday %d: %v", weekday, err)
			continue
		}
		dayEnds, err := time.Parse(time.TimeOnly, day.End)
		if err != nil {
			log.Printf("failed to parse end date for weekday %d: %v", weekday, err)
			continue
		}

		if !dayEnds.After(dayStarts) {
			log.Printf("schedule start time must be greater than end time. start: %s end: %s", dayStarts, dayEnds)
			continue
		}

		diff := dayEnds.Sub(dayStarts)

		parsedSchedule[weekday] = parsedDay{duration: diff}
	}

	var remaining time.Duration

	for {
		if start.After(now) || start.After(end) {
			break
		}

		// In Go weekdays start at sundays, this shifts it to start at mondays
		weekday := (int(start.Weekday()) + 6) % 7
		inExcluded := slices.Contains(excludedDays, start)

		if day, ok := parsedSchedule[weekday]; ok && !inExcluded {
			remaining += day.duration - (time.Duration(s.BreakTime) * time.Second)
		}

		// TODO: dirty
		start = start.Add(24 * time.Hour)
	}

	return int(remaining.Seconds()) - s.DoneSeconds, nil
}
