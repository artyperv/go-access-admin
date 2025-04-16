package access

import (
	"fmt"

	"github.com/foomo/htpasswd"
)

func GetUsers(filePath string) (htpasswd.HashedPasswords, error) {
	// Getting users from .htpasswd
	passwords, err := htpasswd.ParseHtpasswdFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load htpasswd: %w", err)
	}
	return passwords, nil
}

func AddUser(filePath, username, password string) error {
	// Adding or updating user in .htpasswd
	if err := htpasswd.SetPassword(filePath, username, password, htpasswd.HashBCrypt); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}
	return nil
}

func RemoveUser(filePath, username string) error {
	// Deleting user from .htpasswd
	if err := htpasswd.RemoveUser(filePath, username); err != nil {
		return fmt.Errorf("failed to load htpasswd: %w", err)
	}
	return nil
}
