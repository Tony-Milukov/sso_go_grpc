package user

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"log/slog"
	"sso_go_grpc/internal/domain/models"
	"sso_go_grpc/internal/lib/bcrypt"
	"sso_go_grpc/internal/storage"
	"strconv"
	_ "strconv"
)

type StorageInterface interface {
	CreateUser(ctx context.Context, username, email, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uint64) (*models.User, error)
	GetRoleById(ctx context.Context, roleId uint64) (*models.Role, error)
}

type Storage struct {
	StorageInterface
	Db  *sql.DB
	Log *slog.Logger
}

func CreateStorage(db *sql.DB, log *slog.Logger) *Storage {
	return &Storage{Db: db, Log: log}
}

// GetUserByEmail this method gets a user if it not exist it return UserNotExist err
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	op := "storage.postgres.GetUserByEmail"
	log := s.Log.With("op", op)

	var (
		username, hashedPwd, roleName, roleDescription sql.NullString
		userId, roleId                                 sql.NullInt64
	)

	rows, err := s.Db.QueryContext(ctx, `
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
	rows, err := s.Db.QueryContext(ctx, `
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
	rows, err := s.Db.QueryContext(ctx, `
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

	log := s.Log.With("op", op)

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

	err := s.Db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 OR username = $2)", email, username).Scan(&userExists)
	if err != nil {
		log.Debug("Error in Querying user by email or username", err)
		return nil, err
	}

	if userExists {
		return nil, storage.ErrUserExists
	}

	//prepare Db CALL
	prepared, err := s.Db.Prepare("INSERT INTO users(email, password, username) VALUES($1, $2, $3) RETURNING id")

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
