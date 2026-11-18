package domain

import "time"

const MaxHistoryMessages = 10

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
	At      time.Time
}
