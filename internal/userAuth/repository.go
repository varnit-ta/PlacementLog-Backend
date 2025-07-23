package userauth

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/varnit-ta/PlacementLog/internal/db"
	"golang.org/x/crypto/bcrypt"
)

/*
UserAuthRepo handles user authentication data access operations.
Provides methods for user login and registration with database interactions.
*/
type UserAuthRepo struct {
	db *sql.DB
}

/*
NewUserAuthRepo creates a new UserAuthRepo instance with the provided database connection.

Parameters:
- db: The database connection

Returns:
- *UserAuthRepo: A new repository instance
*/
func NewUserAuthRepo(db *sql.DB) *UserAuthRepo {
	return &UserAuthRepo{
		db: db,
	}
}

/*
Login validates user credentials against the database.

Parameters:
- username: The user's username (registration number format: 22bcs1234)
- pass: The user's password

Returns:
- *db.User: The authenticated user information
- error: Any error that occurred during login

The function:
1. Validates the username format using regex (22bcs1234 pattern)
2. Queries the database for the user
3. Compares the provided password with the hashed password using bcrypt
4. Returns user information upon successful authentication

Possible errors:
- "all fields are required": Missing username or password
- "not a valid registration number": Invalid username format
- "no such user exists": User not found in database
- "incorrect password": Password doesn't match
*/
func (repo UserAuthRepo) Login(regno, pass string) (*db.User, error) {
	if regno == "" || pass == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	// Standardize regno to lowercase for case-insensitive login
	regno = strings.ToLower(regno)

	re := regexp.MustCompile(`^\d{2}[a-z]{3}\d{4}$`)

	if !re.MatchString(regno) {
		return nil, fmt.Errorf("not a valid registration number")
	}

	queryString := `
		SELECT id, regno, password, created_at, username
		FROM placement_log_users
		WHERE regno=$1;
	`

	var user db.User
	var hashedPass string

	err := repo.db.QueryRow(queryString, regno).Scan(
		&user.ID,
		&user.Regno,
		&hashedPass,
		&user.CreatedAt,
		&user.Username,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no such user exists")
	}

	if err != nil {
		return nil, fmt.Errorf("db error: %v", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass)); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	return &user, nil
}

/*
Register creates a new user account in the database.

Parameters:
- username: The user's username (registration number format: 22bcs1234)
- pass: The user's password

Returns:
- *db.User: The newly created user information
- error: Any error that occurred during registration

The function:
1. Validates the username format using regex (22bcs1234 pattern)
2. Hashes the password using bcrypt with default cost
3. Inserts the new user into the database
4. Returns the created user information

Possible errors:
- "all fields are required": Missing username or password
- "not a valid registration number": Invalid username format
- "user already exists": Username already taken
- "error hashing pass": Password hashing failed
- "failed to insert user": Database insertion failed
*/
func (repo UserAuthRepo) Register(regno, username, pass string) (*db.User, error) {
	if regno == "" || username == "" || pass == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	// Convert the input to lowercase BEFORE validation
	regno = strings.ToLower(regno)

	re := regexp.MustCompile(`^\d{2}[a-z]{3}\d{4}$`)

	if !re.MatchString(regno) {
		return nil, fmt.Errorf("not a valid registration number")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("error hashing pass: %v", err)
	}

	queryString := `
		INSERT INTO placement_log_users (regno, password, username)
		VALUES ($1, $2, $3)
		RETURNING id, regno, created_at, username;
	`

	var user db.User

	err = repo.db.QueryRow(queryString, regno, hashedPass, username).Scan(
		&user.ID,
		&user.Regno,
		&user.CreatedAt,
		&user.Username,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}

	return &user, nil
}

// Ensure UserAuthRepo implements UserAuthRepository
var _ UserAuthRepository = (*UserAuthRepo)(nil)
