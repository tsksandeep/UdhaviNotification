package udhaviNotification

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

type ChatValueString struct {
	Value string `json:"stringValue"`
}

type ChatMessage struct {
	UserID      ChatValueString `json:"userId"`
	UserName    ChatValueString `json:"username"`
	Body        ChatValueString `json:"body"`
	ChatGroupID ChatValueString `json:"chatGroupId"`
}

type UserData struct {
	UserID    string `json:"userId"`
	ExpoToken string `json:"expoToken"`
}

type ChatGroup struct {
	UserList []UserData `json:"userList"`
}
