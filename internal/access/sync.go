package access

import (
	"fmt"

	"g.pervovsky.ru/go-access-admin/internal/config"
	"g.pervovsky.ru/go-access-admin/internal/storage"
)

func checkAdminInDB(db storage.DB, path string, a config.AdminUser) error {
	// Checks if Admin is in DB and create if not
	exists, err := db.UserExists(a.Username, path)
	if err != nil {
		return err
	}
	if !exists {
		_, err := db.CreateAccess(storage.Access{
			Username:     a.Username,
			Password:     a.Password,
			HtpasswdPath: path,
			IsAdmin:      true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func Synchronize(db storage.DB, config config.Config) error {
	// Synchronizes DB and settings`s Admins with .htpasswds
	for _, f := range config.HtpasswdPaths {
		// 1. Removing users from .htpasswds
		users, err := GetUsers(f.Path)
		if err != nil {
			continue
		}
		for _, u := range users {
			RemoveUser(f.Path, u)
		}

		// 2. Adding Admins to DB
		// File admins
		for _, a := range f.Admins {
			err := checkAdminInDB(db, f.Path, a)
			if err != nil {
				return err
			}
		}

		// Global Admins
		for _, a := range config.Admins {
			err := checkAdminInDB(db, f.Path, a)
			if err != nil {
				return err
			}
		}

		accesses, err := db.GetAccessesByPath(f.Path)
		if err != nil {
			return err
		}

		for _, a := range accesses {
			// 3. Writing db users to .htpasswd
			err = AddUser(f.Path, a.Username, a.Password)
			if err != nil {
				return fmt.Errorf("rewrite htpasswd: %w", err)
			}
		}

	}
	return nil
}
