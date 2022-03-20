package udhaviNotification

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	log "github.com/sirupsen/logrus"
)

const (
	chatGroupCollection = "chatGroup"
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

func getExpoTokenFromSenderIds(ctx context.Context, userId, chatGroupId string) []expo.ExponentPushToken {
	expoTokens := []expo.ExponentPushToken{}

	docSnap, err := firestoreClient.Collection(chatGroupCollection).Doc(chatGroupId).Get(ctx)
	if err != nil {
		log.Errorf("unable to fetch chat group data for %s", chatGroupId)
		return expoTokens
	}

	var chatGroup ChatGroup
	if err := docSnap.DataTo(&chatGroup); err != nil {
		log.Error("unable to unmarshal chat group data")
		return expoTokens
	}

	for _, user := range chatGroup.UserList {
		// Because we should not send notification to the same user
		if user.UserID == userId {
			continue
		}

		token, err := expo.NewExponentPushToken(user.ExpoToken)
		if err != nil {
			log.Errorf("invalid expo token. user id: %s", user.UserID)
			continue
		}

		expoTokens = append(expoTokens, token)
	}

	return expoTokens
}

func PushNotification(ctx context.Context, fsEvent FirestoreEvent) error {
	chatMessage := fsEvent.Value.Fields

	expoTokens := getExpoTokenFromSenderIds(context.Background(), chatMessage.UserID, chatMessage.ChatGroupID)
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
		Priority: expo.DefaultPriority,
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
