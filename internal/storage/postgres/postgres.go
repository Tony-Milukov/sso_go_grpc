package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"log/slog"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/domain/models"
	"sso_go_grpc/internal/lib/bcrypt"
	"sso_go_grpc/internal/storage"
	"strconv"
	_ "strconv"
)

type Storage struct {
	db     *sql.DB
	config *config.Config
	log    *slog.Logger
}

// MustLoad this function returns a Storage, if there is an error , it panics
func MustLoad(cfg *config.Config, log *slog.Logger) *Storage {
	op := "storage.postgres.MustLoad"
	db, err := sql.Open(cfg.DbType, cfg.DbLink)

	if err != nil {
		panic(fmt.Sprintf("%s: %w", op, err))
	}

	fmt.Printf("Database was succesfully connected\n")

	return &Storage{db: db, log: log}
}

// GetUserByEmail this method gets a user if it not exist it return UserNotExist err
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	op := "storage.postgres.GetUserByEmail"
	log := s.log.With("op", op)

	var (
		username, hashedPwd, roleName, roleDescription sql.NullString
		userId, roleId                                 sql.NullInt64
	)

	rows, err := s.db.QueryContext(ctx, `
        SELECT u.id, u.username, u.password, r.name, r.id, r.description
        FROM users u
        LEFT JOIN "userRoles" ur ON u.id = ur.userId
        LEFT JOIN roles r ON ur.roleId = r.id
		WHERE email = $1
`, email)

	if err != nil {
		log.Debug("Error by Getting user by email in the database")
		return nil, err
	}

	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		err := rows.Scan(&userId, &username, &hashedPwd, &roleName, &roleId, &roleDescription)
		if err != nil {
			return nil, err
		}
		if roleId.Valid {
			roles = append(roles, &models.Role{Id: uint64(roleId.Int64), Name: roleName.String, Description: roleDescription.String})
		}
	}

	if !userId.Valid {
		return nil, storage.ErrUserNotExists
	}

	return &models.User{Email: email, Username: username.String, UserId: uint64(userId.Int64), Password: hashedPwd.String, Roles: roles}, nil
}

// GetUserByUsername this method gets a user if it not exist it return UserNotExist err
func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {

	// user email and hashed password
	var email, hashedPwd, roleName, roleDescription sql.NullString
	var userId, roleId sql.NullInt64

	//create and execute sql
	rows, err := s.db.QueryContext(ctx, `
        SELECT u.username, u.email, u.password, r.name, r.id, r.description
        FROM users u
        LEFT JOIN "userRoles" ur ON u.id = ur.userId
        LEFT JOIN roles r ON ur.roleId = r.id
		                            WHERE username = $1
`, username)

	// if there was an error in sql
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		if err = rows.Scan(&userId, &email, &hashedPwd, &roleId, &roleName, &roleDescription); err != nil {
			return nil, err
		}

		if !userId.Valid {
			return nil, storage.ErrUserNotExists
		}

		if roleId.Valid {
			roles = append(roles, &models.Role{Id: uint64(roleId.Int64), Description: roleDescription.String, Name: roleName.String})
		}
	}
	//if there was an error in converting string to uint64
	if err != nil {
		return nil, err
	}

	return &models.User{Email: email.String, Username: username, UserId: uint64(userId.Int64)}, nil
}

func (s *Storage) GetUserById(ctx context.Context, userId uint64) (*models.User, error) {
	var (
		email, username, hashedPwd string
		userFound                  bool
	)
	rows, err := s.db.QueryContext(ctx, `
        SELECT u.username, u.email, u.password, r.name, r.id, r.description
        FROM users u
        LEFT JOIN "userRoles" ur ON u.id = ur.userId
        LEFT JOIN roles r ON ur.roleId = r.id
        WHERE u.id = $1`, userId)
	if err != nil {
		return nil, err
	}

	var roles []*models.Role

	for rows.Next() {
		userFound = true
		var (
			roleId sql.NullInt64
			roleName,
			roleDescription sql.NullString
		)

		if err := rows.Scan(&username, &email, &hashedPwd, &roleName, &roleId, &roleDescription); err != nil {
			if errors.Is(sql.ErrNoRows, err) {
				return nil, storage.ErrUserNotExists
			}
			return nil, err
		}
		if roleId.Valid {
			roles = append(roles, &models.Role{Id: uint64(roleId.Int64), Name: roleName.String, Description: roleDescription.String})
		}
	}

	if !userFound {
		return nil, storage.ErrUserNotExists
	}
	defer rows.Close()
	return &models.User{Email: email, Username: username, UserId: userId, Roles: roles}, nil
}

// CreateUser this method creates new user and proofs if user with that email or username does exist
func (s *Storage) CreateUser(ctx context.Context, email, password, username string) (*models.User, error) {
	op := "app.grpc.app"

	log := s.log.With("op", op)

	// userId as string
	var userIdStr string
	var userExists bool
	// check if user with that email exists
	if _, err := s.GetUserByEmail(ctx, email); err == nil {
		log.Error("User with that username already exists", err)
		return nil, storage.ErrUserExists
	}

	// check if user with that username exists
	if _, err := s.GetUserByUsername(ctx, username); err == nil {
		log.Error("User with that username already exists", err)
		return nil, storage.ErrUserExists
	}

	err := s.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 OR username = $2)", email, username).Scan(&userExists)
	if err != nil {
		log.Debug("Error in Querying user by email or username", err)
		return nil, err
	}

	if userExists {
		return nil, storage.ErrUserExists
	}

	//prepare DB CALL
	prepared, err := s.db.Prepare("INSERT INTO users(email, password, username) VALUES($1, $2, $3) RETURNING id")

	//if there was an error in preparing sql
	if err != nil {
		log.Debug("Error in preparing sql", err)
		return nil, err
	}

	pwdHash, err := bcrypt.HashPassword(password)

	//if err in hashing password
	if err != nil {
		log.Error("Error on hashing password", err)
		return nil, err
	}

	//execute sql
	err = prepared.QueryRowContext(ctx, email, pwdHash, username).Scan(&userIdStr)

	//if there was an error in executing sql
	if err != nil {
		log.Error("Error on executing sql", err)
		return nil, err
	}

	// str to uint64
	userId, err := strconv.ParseUint(userIdStr, 10, 64)

	//if there was an error in converting string to uint64
	if err != nil {
		log.Error("Error on converting string to uint64", err)
		return nil, err
	}

	user, err := s.GetUserById(ctx, userId)

	//fill the user model and return it
	return user, nil
}

// CreateRole this creates a new Role in the database
func (s *Storage) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	op := "storage.postgres.CreateRole"

	_, err := s.GetRoleByName(ctx, name)

	if err == nil {
		return nil, storage.ErrRoleExists
	}

	//setting up logger
	logger := s.log.With("op", op)

	//the new role ID
	var roleId sql.NullInt64

	//prepare sql call to create new role
	prepared, err := s.db.Prepare(`INSERT INTO roles(name, description)  VALUES ($1, $2) RETURNING id`)

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
	err := s.db.QueryRowContext(ctx, "SELECT name, description FROM roles r WHERE r.id = $1", id).Scan(&name, &description)

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
	err := s.db.QueryRowContext(ctx, "SELECT id, description FROM roles r WHERE r.name = $1", name).Scan(&id, &description)

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
	// prepare SQL to delete a role from the DB
	prepared, err := s.db.Prepare(`DELETE FROM roles r WHERE r.id = $1`)

	// close connection in the end
	defer prepared.Close()

	// if there is an error in preparing
	if err != nil {
		return err
	}

	// execute the prepared
	//get the result context
	result, err := prepared.ExecContext(ctx, roleId)

	if err != nil {
		return err
	}

	//get the affected rows
	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		return storage.ErrRoleNotExists
	}

	//if there were no errors
	return nil
}

func (s *Storage) UpdateRole(ctx context.Context, name, description string, roleId uint64) (*models.Role, error) {
	op := "storage.postgres.UpdateRole"
	logger := s.log.With("op", op)

	prepared, err := s.db.Prepare(`UPDATE roles r SET name = $1, description = $2 WHERE r.id = $3`)

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
	logger := s.log.With("op", op)

	prepared, err := s.db.Prepare(`INSERT INTO "userRoles" (userId, roleId) VALUES($1,$2)`)

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
	logger := s.log.With("op", op)

	prepared, err := s.db.Prepare(`DELETE FROM "userRoles" ur WHERE ur.userId = $1 AND ur.roleId = $2 `)

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

	err := s.db.QueryRowContext(ctx, `SELECT id FROM "userRoles" WHERE roleId = $1 AND userId = $2`, roleId, userId).Scan(&userRoleId)

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
