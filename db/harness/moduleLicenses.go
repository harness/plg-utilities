package harness

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	moduleLicenseCollection = "moduleLicenses"
)

type ModuleLicenseDAO struct {
	ModuleLicenseCollection *mongo.Collection
}

func NewModuleLicenseDAO(mongoDb *mongo.Database) (*ModuleLicenseDAO, error) {
	dao := &ModuleLicenseDAO{
		ModuleLicenseCollection: mongoDb.Collection(moduleLicenseCollection),
	}
	return dao, nil
}

func (a *ModuleLicenseDAO) ListWithCursor(ctx context.Context) (*mongo.Cursor, error) {
	cursor, err := a.ModuleLicenseCollection.Find(ctx, bson.D{})
	if err != nil {
		logrus.WithError(err).Errorf("failed to retrieve documents from collection %s", accountsCollection)
		return nil, err
	}
	logrus.Infof("sucessfully retrieved collection list with cursor %s", a.ModuleLicenseCollection.Name())
	return cursor, nil
}
