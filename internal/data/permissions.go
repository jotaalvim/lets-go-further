package data

import (
	"context"
	"database/sql"
	"slices"
	"time"
)

type Permissions []string

type PermissionsModel struct {
	DB *sql.DB
}

func (p Permissions) Includes(code string) bool {
	return slices.Contains(p, code)
}

// GetAllForUser returns all permission codes for a specific user
func (m *PermissionsModel) GetAllForUser(userID int) (Permissions, error) {

	query := `
	SELECT permissions.code 
	FROM permissions
	INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
	INNER JOIN users ON users_permissions.user_id = users.id 
	WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var permissions Permissions

	for rows.Next() {
		var perm string
		err := rows.Scan(&perm)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return permissions, nil

}
