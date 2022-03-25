package chatNotification

import (
	"time"
)

type UpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask UpdateMask     `json:"updateMask"`
}

type FirestoreValue struct {
	CreateTime time.Time   `json:"createTime"`
	Fields     ChatMessage `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

type ChatMessage struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	Body        string `json:"body"`
	ChatGroupID string `json:"chatGroupId"`
}

type UserData struct {
	UserID    string `json:"userId"`
	ExpoToken string `json:"expoToken"`
}

type ChatGroup struct {
	UserList []UserData `json:"userList"`
}

type Notification struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Timestamp time.Time `json:"timestamp"`
}
