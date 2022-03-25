package chatNotification

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	log "github.com/sirupsen/logrus"
)

const (
	chatGroupCollection    = "chatGroup"
	notificationCollection = "notifications"

	alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	expoClient      *expo.PushClient
	firestoreClient *firestore.Client
	firebaseConfig  *firebase.Config
)

func init() {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{log.FieldKeyMsg: "message"},
	})
	log.SetLevel(log.InfoLevel)

	firebaseConfig = &firebase.Config{
		ProjectID:   "udhavi-dev",
		DatabaseURL: "https://udhavi-dev.firebaseio.com",
	}

	ctx := context.Background()

	firebaseApp, err := firebase.NewApp(ctx, firebaseConfig)
	if err != nil {
		log.Errorf("initializing firebase app: %s", err)
		return
	}

	firestoreClient, err = firebaseApp.Firestore(ctx)
	if err != nil {
		log.Errorf("initializing firestore client: %s", err)
		return
	}

	expoClient = expo.NewPushClient(nil)
}

func getDocId() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	for i, byt := range b {
		b[i] = alphanum[int(byt)%len(alphanum)]
	}

	return string(b)
}

func getExpoTokenFromSenderIds(ctx context.Context, chatMessage ChatMessage) []expo.ExponentPushToken {
	expoTokens := []expo.ExponentPushToken{}

	docSnap, err := firestoreClient.Collection(chatGroupCollection).Doc(chatMessage.ChatGroupID).Get(ctx)
	if err != nil {
		log.Errorf("unable to fetch chat group data for %s", chatMessage.ChatGroupID)
		return expoTokens
	}

	var chatGroup ChatGroup
	if err := docSnap.DataTo(&chatGroup); err != nil {
		log.Error("unable to unmarshal chat group data")
		return expoTokens
	}

	for _, user := range chatGroup.UserList {
		// Because we should not send notification to the same user
		if user.UserID == chatMessage.UserID {
			continue
		}

		token, err := expo.NewExponentPushToken(user.ExpoToken)
		if err != nil {
			log.Errorf("invalid expo token. user id: %s", user.UserID)
			continue
		}

		docId := getDocId()
		if docId == "" {
			log.Errorf("unable to generate doc id...")
			continue
		}

		firestoreClient.Collection(notificationCollection).Doc(user.UserID).Collection("list").Doc(docId).Create(ctx, Notification{
			ID:        docId,
			Body:      chatMessage.Body,
			Title:     "New Message from " + chatMessage.UserName,
			Category:  "chat",
			Timestamp: time.Now(),
		})

		expoTokens = append(expoTokens, token)
	}

	return expoTokens
}

func PushNotification(ctx context.Context, fsEvent FirestoreEvent) error {
	chatMessage := fsEvent.Value.Fields

	expoTokens := getExpoTokenFromSenderIds(context.Background(), chatMessage)
	if len(expoTokens) == 0 {
		errMsg := "no expo tokens to send notification"
		log.Error(errMsg)
		return errors.New(errMsg)
	}

	pushMessage := &expo.PushMessage{
		To:       expoTokens,
		Body:     chatMessage.Body,
		Sound:    "default",
		Title:    "New Message from " + chatMessage.UserName,
		Priority: expo.HighPriority,
		Data: map[string]string{
			"category": "chat",
		},
	}

	response, err := expoClient.Publish(pushMessage)
	if err != nil {
		log.Error(err)
		return err
	}

	if response.ValidateResponse() != nil {
		log.Error(response.PushMessage.To, "failed")
		return nil
	}

	return nil
}
