package mq

// Logical topic keys. These are the keys used in config/config.yaml under
// message_publishers/message_consumers, and the map keys internal/server
// uses when matching a topic to its consumer handler.
const (
	TopicUserCreated = "user_created"
	TopicUserDeleted = "user_deleted"
)
