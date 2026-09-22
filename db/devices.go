package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
)

func (s *Store) AddDevice(ctx context.Context, device *Device) error {
	query := `
        INSERT INTO devices (name, secret_key)
        VALUES (?,?)
    `

	secret := randomKey(12)
	res, err := s.db.ExecContext(ctx, query, device.Name, secret)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	device.ID = int(id)
	device.SecretKey = secret

	return nil
}

func (s *Store) ListDevices(ctx context.Context) ([]*Device, error) {
	query := `
		SELECT id, name, secret_key
		FROM devices
		`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*Device, 0)
	for rows.Next() {
		dev := &Device{}
		err := rows.Scan(&dev.ID, &dev.Name, &dev.SecretKey)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}

	return devices, rows.Err()
}

func (s *Store) FindDeviceByID(ctx context.Context, id int) (*Device, error) {
	query := `
		SELECT id, name, secret_key
		FROM devices
		WHERE id = ? LIMIT 1`

	var dev Device
	err := s.db.QueryRowContext(ctx, query, id).Scan(&dev.ID, &dev.Name, &dev.SecretKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no device found based on id: %d", id)
		}
		return nil, err
	}

	return &dev, nil
}

func (s *Store) UpdateDevice(ctx context.Context, id int, device Device) error {
	query := `
        UPDATE devices
		SET name = ?,
   			secret_key = ?
		WHERE id = ?

    `
	_, err := s.db.ExecContext(ctx, query, device.Name, device.SecretKey, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteDeviceByID(ctx context.Context, id int) error {
	query := `
		DELETE FROM devices
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func randomKey(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
