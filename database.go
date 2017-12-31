package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// DB contains the DB methods needed by application
type DB interface {
	Save(data *ChannelData) error
	Get() ([]*ChannelData, error)
}

// ChannelData contains the data to be saved
type ChannelData struct {
	ID          string
	Region      string
	ReportRunes bool
	Summoners   map[string]Player
	RuneDesc    string
}

// FireBaseDB contains the connection to FireBase
type FireBaseDB struct {
	client     *firestore.Client
	collection string
}

// New creates a new FireBase connection
func New(collectionName string) *FireBaseDB {
	fb := FireBaseDB{collection: collectionName}
	opt := option.WithCredentialsFile("fbaccountkey.json")
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	fb.client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	return &fb
}

// Save saves the data of the channel
func (fb *FireBaseDB) Save(data *ChannelData) error {
	_, err := fb.client.Collection(fb.collection).Doc(data.ID).Set(context.Background(), data)
	if err != nil {
		return fmt.Errorf("Failed to add data: %v\n", err)
	}
	return nil
}

// Get returns all data
func (fb *FireBaseDB) Get() ([]*ChannelData, error) {
	iter := fb.client.Collection(fb.collection).Documents(context.Background())
	var allData []*ChannelData
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over collection: %v\n", err)
		}
		var data ChannelData
		if err := doc.DataTo(&data); err != nil {
			return nil, fmt.Errorf("failed to unmarshall data: %v\n", err)
		}
		allData = append(allData, &data)
	}
	return allData, nil
}
