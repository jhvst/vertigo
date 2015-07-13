// Package gorm is PostgreSQL and SQLite database access layer (DAL) for users and posts.
//
// The package contains three files:
// * connection.go, which handles the actual database connection as singleton
// * posts.go, which handles CRUD methods for posts
// * users.go, which handles CRUD methods for users
//
// All methods defined in posts.go and users.go should be implemented in other drivers as well,
// unless specifically said otherwise.
package gorm