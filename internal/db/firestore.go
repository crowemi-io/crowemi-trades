package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	CollectionAccounts         = "accounts"
	CollectionActivities       = "activities"
	CollectionCorporateActions = "corporate_actions"
	CollectionOrders           = "orders"
	CollectionPortfolios       = "portfolios"
	CollectionPositions        = "positions"
	CollectionWatchlists       = "watchlists"
	CollectionWithholdings     = "withholdings"
)

type Firestore struct {
	Client *firestore.Client
}

type Document interface {
	GetID() string
	SetID(id string)
}

func NewFirestore(client *firestore.Client) *Firestore {
	return &Firestore{Client: client}
}

func Create[T Document](ctx context.Context, fs *Firestore, collection string, doc T) (string, error) {
	if fs == nil || fs.Client == nil {
		return "", errors.New("firestore client is not initialized")
	}

	col := fs.Client.Collection(collection)
	id := doc.GetID()

	var ref *firestore.DocumentRef
	if id == "" {
		ref = col.NewDoc()
	} else {
		ref = col.Doc(id)
	}

	if _, err := ref.Set(ctx, doc); err != nil {
		return "", err
	}

	doc.SetID(ref.ID)
	return ref.ID, nil
}

func Get[T Document](ctx context.Context, fs *Firestore, collection, id string) (T, error) {
	var zero T
	if fs == nil || fs.Client == nil {
		return zero, errors.New("firestore client is not initialized")
	}

	doc, err := newDocument[T]()
	if err != nil {
		return zero, err
	}

	snapshot, err := fs.Client.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		return zero, err
	}
	if err := snapshot.DataTo(doc); err != nil {
		return zero, err
	}

	doc.SetID(snapshot.Ref.ID)
	return doc, nil
}

func List[T Document](ctx context.Context, fs *Firestore, collection string) ([]T, error) {
	if fs == nil || fs.Client == nil {
		return nil, errors.New("firestore client is not initialized")
	}

	iter := fs.Client.Collection(collection).Documents(ctx)
	defer iter.Stop()

	var docs []T
	for {
		snapshot, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		doc, err := newDocument[T]()
		if err != nil {
			return nil, err
		}
		if err := snapshot.DataTo(doc); err != nil {
			return nil, err
		}

		doc.SetID(snapshot.Ref.ID)
		docs = append(docs, doc)
	}

	return docs, nil
}

func Update[T Document](ctx context.Context, fs *Firestore, collection string, doc T) error {
	if fs == nil || fs.Client == nil {
		return errors.New("firestore client is not initialized")
	}

	id := doc.GetID()
	if id == "" {
		return errors.New("document id is required for update")
	}

	_, err := fs.Client.Collection(collection).Doc(id).Set(ctx, doc, firestore.MergeAll)
	return err
}

func Delete(ctx context.Context, fs *Firestore, collection, id string) error {
	if fs == nil || fs.Client == nil {
		return errors.New("firestore client is not initialized")
	}

	if id == "" {
		return errors.New("document id is required for delete")
	}

	_, err := fs.Client.Collection(collection).Doc(id).Delete(ctx)
	return err
}

func newDocument[T Document]() (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType == nil {
		return zero, errors.New("document type is nil")
	}

	if targetType.Kind() != reflect.Ptr {
		return zero, fmt.Errorf("document type %s must be a pointer", targetType)
	}

	value := reflect.New(targetType.Elem())
	doc, ok := value.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("unable to build document for type %s", targetType)
	}

	return doc, nil
}
