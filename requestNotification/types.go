package requestNotification

import (
	"time"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type ExpoNotification struct {
	Title      string
	Body       string
	ExpoTokens []expo.ExponentPushToken
}

type UpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask UpdateMask     `json:"updateMask"`
}

type FirestoreValue struct {
	CreateTime time.Time `json:"createTime"`
	Fields     Request   `json:"fields"`
	Name       string    `json:"name"`
	UpdateTime time.Time `json:"updateTime"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Request struct {
	Category             string    `json:"category"`
	Date                 time.Time `json:"date"`
	DeliveryTime         time.Time `json:"deliveryTime"`
	ID                   string    `json:"id"`
	RequestorID          string    `json:"requestorId"`
	Info                 string    `json:"info"`
	Location             Location  `json:"location"`
	Name                 string    `json:"name"`
	Notes                string    `json:"notes"`
	PhoneNumber          string    `json:"phoneNumber"`
	RequestorPhoneNumber string    `json:"requestorPhoneNumber"`
	Status               string    `json:"status"`
	AssignedVolunteerIds []string  `json:"assignedVolunteerIds"`
}

type UserData struct {
	Name        string `json:"name"`
	ExpoToken   string `json:"expoToken"`
	PhoneNumber string `json:"phoneNumber"`
}

type Notification struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Timestamp time.Time `json:"timestamp"`
}
