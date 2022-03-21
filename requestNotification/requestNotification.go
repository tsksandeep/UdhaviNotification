package requestNotification

import (
	"context"
	"path/filepath"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	log "github.com/sirupsen/logrus"
)

const (
	userCollection = "users"
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

func getUserData(ctx context.Context, id string) *UserData {
	docSnap, err := firestoreClient.Collection(userCollection).Doc(id).Get(ctx)
	if err != nil {
		log.Errorf("unable to fetch user data for %s", id)
		return nil
	}

	var userData UserData
	if err := docSnap.DataTo(&userData); err != nil {
		log.Error("unable to unmarshal user data")
		return nil
	}

	return &userData
}

func assignedVolunteerNotification(ctx context.Context, request Request) []ExpoNotification {
	expoNotifications := []ExpoNotification{}

	requestorData := getUserData(ctx, request.RequestorID)
	if requestorData == nil {
		return expoNotifications
	}

	for _, volunteerId := range request.AssignedVolunteerIds {
		volunteerData := getUserData(ctx, volunteerId)
		if volunteerData == nil {
			continue
		}

		token, err := expo.NewExponentPushToken(requestorData.ExpoToken)
		if err != nil {
			log.Errorf("invalid expo token. user id: %s", request.RequestorID)
			continue
		}

		expoNotifications = append(expoNotifications, ExpoNotification{
			Title:      "Update for Request - " + request.ID,
			Body:       volunteerData.Name + " has been assigned for your request",
			ExpoTokens: []expo.ExponentPushToken{token},
		})
	}

	return expoNotifications
}

func pushExpoNotificationGoRoutine(waitGroup *sync.WaitGroup, expoNotification ExpoNotification) {
	defer waitGroup.Done()

	pushMessage := &expo.PushMessage{
		To:       expoNotification.ExpoTokens,
		Body:     expoNotification.Body,
		Sound:    "default",
		Title:    expoNotification.Title,
		Priority: expo.DefaultPriority,
	}

	response, err := expoClient.Publish(pushMessage)
	if err != nil {
		log.Error(err)
		return
	}

	if response.ValidateResponse() != nil {
		log.Error(response.PushMessage.To, "failed")
	}
}

func pushExpoNotifications(expoNotifications []ExpoNotification) {
	if len(expoNotifications) == 0 {
		log.Error("no expo notifications to send")
		return
	}

	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(expoNotifications))

	for _, expoNotification := range expoNotifications {
		go pushExpoNotificationGoRoutine(waitGroup, expoNotification)
	}

	waitGroup.Wait()
}

func PushNotification(ctx context.Context, fsEvent FirestoreEvent) error {
	request := fsEvent.Value.Fields
	updatedFieldPaths := fsEvent.UpdateMask.FieldPaths

	for _, updatedFieldPath := range updatedFieldPaths {
		field := filepath.Base(updatedFieldPath)

		switch field {
		// TODO: handle more cases
		case "assignedVolunteerIds":
			expoNotifications := assignedVolunteerNotification(ctx, request)
			pushExpoNotifications(expoNotifications)
		}
	}

	return nil
}
