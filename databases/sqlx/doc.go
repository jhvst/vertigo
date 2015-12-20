// Package sqlx is PostgreSQL and SQLite database access layer (DAL) for users and posts.
//
// The package contains three files:
// * connection.go, which handles the actual database connection as singleton
// * posts.go, which handles CRUD methods for posts
// * users.go, which handles CRUD methods for users
// * email.go, which handles method for sending email to users
// * settings.go, which handles CU methods for settings
//
// All methods defined this package should be implemented in other drivers as well,
// unless specifically said otherwise.
package sqlx
