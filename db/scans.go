package db

import (
	"context"
)

func (s *Store) NewScan(ctx context.Context, scan *Scan) error {
	query := `INSERT INTO scans (uid, student_id) VALUES (?,?)`

	res, err := s.db.ExecContext(ctx, query, scan.UID, scan.StudentID)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	scan.ID = int(id)
	return nil
}

func (s *Store) ListScans(ctx context.Context) ([]*Scan, error) {
	query := `SELECT id, uid, timestamp, student_id FROM scans`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scans := make([]*Scan, 0)
	for rows.Next() {
		sc := &Scan{}
		err := rows.Scan(&sc.ID, &sc.UID, &sc.Timestamp, &sc.StudentID)
		if err != nil {
			return nil, err
		}
		scans = append(scans, sc)
	}

	return scans, rows.Err()
}
