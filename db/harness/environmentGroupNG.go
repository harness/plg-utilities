package harness

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	environmentGroupNGCollection = "environmentGroupNG"
)

type EnvironmentGroupNGDAO struct {
	EnvironmentGroupNGCollection *mongo.Collection
}

func NewEnvironmentGroupNGDAO(mongoDb *mongo.Database) (*EnvironmentGroupNGDAO, error) {
	dao := &EnvironmentGroupNGDAO{
		EnvironmentGroupNGCollection: mongoDb.Collection(environmentGroupNGCollection),
	}
	return dao, nil
}

//func (a *AccountDAO) Create(ctx context.Context, account *core.Account) (*core.Account, error) {
//	_, err := a.EnvironmentGroupNGDAO.InsertOne(ctx, account)
//	if err != nil {
//		logrus.WithError(err).Errorf("failed to insert document in Mongo for account %s", account.GetAccountIdAsString())
//		return nil, fmt.Errorf("failed to create application %s", account.GetAccountIdAsString())
//	}
//	return account, nil
//}
//
//func (a *AccountDAO) Delete(ctx context.Context, account *core.Account) (*core.Account, error) {
//	//todo
//	return nil, nil
//}
//
//func (a *AccountDAO) Update(ctx context.Context, account *core.Account) (*core.Account, error) {
//	//todo
//	return nil, nil
//}

func (a *EnvironmentGroupNGDAO) ListWithCursor(ctx context.Context) (*mongo.Cursor, error) {
	cursor, err := a.EnvironmentGroupNGCollection.Find(ctx, bson.D{})
	if err != nil {
		logrus.WithError(err).Errorf("failed to retrieve documents from collection %s", environmentGroupNGCollection)
		return nil, err
	}
	logrus.Infof("sucessfully retrieved collection list with cursor %s", a.EnvironmentGroupNGCollection.Name())
	return cursor, nil
}
