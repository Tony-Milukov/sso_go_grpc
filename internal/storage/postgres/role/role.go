package role

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sso_go_grpc/internal/domain/models"
	"sso_go_grpc/internal/storage"
)

type StorageInterface interface {
	CreateRole(ctx context.Context, name, description string) (*models.Role, error)
	UpdateRole(ctx context.Context, name, description string, roleId uint64) (*models.Role, error)
	DeleteRole(ctx context.Context, roleId uint64) error
}

type Storage struct {
	StorageInterface
	Db  *sql.DB
	Log *slog.Logger
}

func CreateStorage(db *sql.DB, log *slog.Logger) *Storage {
	return &Storage{Db: db, Log: log}
}

// CreateRole this creates a new Role in the database
func (s *Storage) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	op := "storage.postgres.CreateRole"

	_, err := s.GetRoleByName(ctx, name)

	if err == nil {
		return nil, storage.ErrRoleExists
	}

	//setting up logger
	logger := s.Log.With("op", op)

	//the new role ID
	var roleId sql.NullInt64

	//prepare sql call to create new role
	prepared, err := s.Db.Prepare(`INSERT INTO roles(name, description)  VALUES ($1, $2) RETURNING id`)

	//if there was an error in preparing sql
	if err != nil {
		logger.Debug("Error in preparing sql", err)
		return nil, err
	}

	//call the prepared request
	//if there ws an error return it
	if err = prepared.QueryRowContext(ctx, name, description).Scan(&roleId); err != nil {
		return nil, err
	}

	//get the new role by its ID
	role, err := s.GetRoleById(ctx, uint64(roleId.Int64))

	//if there ws an error return it
	if err != nil {
		return nil, err
	}

	//return new role
	return role, nil
}

// GetRoleById is getting a role by id and returns &models.Role
func (s *Storage) GetRoleById(ctx context.Context, id uint64) (*models.Role, error) {
	//role params
	var description, name *sql.NullString

	//check if there is
	//sql call to get the information
	err := s.Db.QueryRowContext(ctx, "SELECT name, description FROM roles r WHERE r.id = $1", id).Scan(&name, &description)

	//handle error
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, storage.ErrRoleNotExists
		}
		return nil, err
	}

	if !name.Valid {
		return nil, storage.ErrRoleNotExists
	}

	//return the Role model
	return &models.Role{Id: id, Description: description.String, Name: name.String}, nil
}

// GetRoleByName is getting a role by name and returns &models.Role
func (s *Storage) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {

	//role params
	var (
		description *sql.NullString
		id          *sql.NullInt64
	)

	//sql call to get the information
	err := s.Db.QueryRowContext(ctx, "SELECT id, description FROM roles r WHERE r.name = $1", name).Scan(&id, &description)

	//handle error
	if err != nil {
		return nil, err
	}

	if !id.Valid {
		return nil, storage.ErrRoleNotExists
	}

	//return the Role model
	return &models.Role{Id: uint64(id.Int64), Description: description.String, Name: name}, nil
}

func (s *Storage) DeleteRole(ctx context.Context, roleId uint64) error {
	tx, err := s.Db.BeginTx(ctx, nil)

	// if there was an error on creating transaction
	if err != nil {
		return err
	}

	// delete the role from all users
	if _, err = tx.ExecContext(ctx, `DELETE FROM "userRoles" ur WHERE ur.roleId = $1`, roleId); err != nil {
		tx.Rollback()
		return err
	}

	// delete the role
	if _, err = tx.ExecContext(ctx, `DELETE FROM roles r WHERE r.id = $1`, roleId); err != nil {
		tx.Rollback()
		return err
	}

	//commit the changes to the database
	if err = tx.Commit(); err != nil {
		return err
	}

	//if there were no errors
	return nil
}

func (s *Storage) UpdateRole(ctx context.Context, name, description string, roleId uint64) (*models.Role, error) {
	op := "storage.postgres.UpdateRole"
	logger := s.Log.With("op", op)
	prepared, err := s.Db.Prepare(`UPDATE roles r SET name = $1, description = $2 WHERE r.id = $3`)

	if err != nil {
		return nil, err
	}

	defer prepared.Close()

	_, err = prepared.ExecContext(ctx, name, description, roleId)

	if err != nil {
		logger.Debug("Error  On executing query", err)
		return nil, err
	}

	return s.GetRoleById(ctx, roleId)
}

func (s *Storage) AddUserRole(
	ctx context.Context,
	roleId,
	userId uint64,
) error {
	op := "storage.postgres.AddUserRole"
	logger := s.Log.With("op", op)

	prepared, err := s.Db.Prepare(`INSERT INTO "userRoles" (userId, roleId) VALUES($1,$2)`)

	if err != nil {
		logger.Debug("Error on preparing the query")
		return err
	}

	_, err = prepared.ExecContext(ctx, userId, roleId)

	if err != nil {
		logger.Debug("Error on executing query")
		return err
	}

	defer prepared.Close()

	return nil
}

func (s *Storage) RemoveUserRole(ctx context.Context, roleId, userId uint64) error {
	op := "storage.postgres.RemoveUserRole"
	logger := s.Log.With("op", op)

	prepared, err := s.Db.Prepare(`DELETE FROM "userRoles" ur WHERE ur.userId = $1 AND ur.roleId = $2 `)

	if err != nil {
		logger.Debug("Error on preparing the query")
		return err
	}

	defer prepared.Close()

	result, err := prepared.ExecContext(ctx, userId, roleId)

	deletedRows, err := result.RowsAffected()

	if err != nil {
		logger.Debug("Error on Getting deletedRows count")
		return err
	}

	if deletedRows == 0 {
		return storage.ErrNoDelete
	}

	return nil
}

func (s *Storage) VerifyUserRole(ctx context.Context, roleId, userId uint64) (bool, error) {
	var userRoleId sql.NullInt64

	err := s.Db.QueryRowContext(ctx, `SELECT id FROM "userRoles" WHERE roleId = $1 AND userId = $2`, roleId, userId).Scan(&userRoleId)

	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}

	if !userRoleId.Valid {
		return false, nil
	}

	return true, nil
}
