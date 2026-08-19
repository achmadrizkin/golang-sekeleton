package database

const (
	// UserTableName is the table name for users in the database.
	UserTableName = "users"
	// UserFields is the ordered column list used by every user query, kept
	// in one place so INSERT/SELECT/Scan stay in sync.
	UserFields = "id, username, email, password, full_name, avatar, role, is_active, email_verified, created_at, updated_at"
)
