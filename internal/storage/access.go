package storage

import (
	"time"
)

type Access struct {
	ID           int
	Username     string
	Password     string
	HtpasswdPath string
	ExpiresAt    time.Time
	IsAdmin      bool
}

func (d *DB) CreateAccess(a Access) (int64, error) {
	result, err := d.conn.Exec(`
        INSERT INTO accesses (username, password, htpasswd_path, expires_at, is_admin)
        VALUES (?, ?, ?, ?, ?)`,
		a.Username, a.Password, a.HtpasswdPath, a.ExpiresAt, a.IsAdmin,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (d *DB) DeleteAccess(id int64) error {
	_, err := d.conn.Exec(`DELETE FROM accesses WHERE id = ? AND is_admin = 0`, id)
	return err
}

func (d *DB) GetAllAccesses() ([]Access, error) {
	rows, err := d.conn.Query(`SELECT id, username, password, htpasswd_path, expires_at, is_admin FROM accesses`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []Access
	for rows.Next() {
		var a Access
		err := rows.Scan(&a.ID, &a.Username, &a.Password, &a.HtpasswdPath, &a.ExpiresAt, &a.IsAdmin)
		if err != nil {
			return nil, err
		}
		accesses = append(accesses, a)
	}

	return accesses, nil
}

func (d *DB) GetExpiredAccesses() ([]Access, error) {
	now := time.Now()
	rows, err := d.conn.Query(`
        SELECT id, username, password, htpasswd_path, expires_at, is_admin
        FROM accesses
        WHERE expires_at < ? AND is_admin = 0`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []Access
	for rows.Next() {
		var a Access
		err := rows.Scan(&a.ID, &a.Username, &a.Password, &a.HtpasswdPath, &a.ExpiresAt, &a.IsAdmin)
		if err != nil {
			return nil, err
		}
		expired = append(expired, a)
	}

	return expired, nil
}

func (d *DB) GetAccessesByPath(htpasswdPath string) ([]Access, error) {
	rows, err := d.conn.Query(`
        SELECT id, username, password, htpasswd_path, expires_at, is_admin
        FROM accesses
        WHERE htpasswd_path = ?`, htpasswdPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Access
	for rows.Next() {
		var a Access
		err := rows.Scan(&a.ID, &a.Username, &a.Password, &a.HtpasswdPath, &a.ExpiresAt, &a.IsAdmin)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}

	return result, nil
}

func (d *DB) UserExists(username, htpasswdPath string) (bool, error) {
	var exists bool
	err := d.conn.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM accesses WHERE username = ? AND htpasswd_path = ?
        )`, username, htpasswdPath,
	).Scan(&exists)
	return exists, err
}
