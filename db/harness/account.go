package harness

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	accountsCollection = "accounts"
)

type AccountDAO struct {
	AccountCollection *mongo.Collection
}

func NewAccountDAO(mongoDb *mongo.Database) (*AccountDAO, error) {
	dao := &AccountDAO{
		AccountCollection: mongoDb.Collection(accountsCollection),
	}
	return dao, nil
}

//func (a *AccountDAO) Create(ctx context.Context, account *core.Account) (*core.Account, error) {
//	_, err := a.AccountCollection.InsertOne(ctx, account)
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

func (a *AccountDAO) ListWithCursor(ctx context.Context) (*mongo.Cursor, error) {
	cursor, err := a.AccountCollection.Find(ctx, bson.D{})
	if err != nil {
		logrus.WithError(err).Errorf("failed to retrieve documents from collection %s", accountsCollection)
		return nil, err
	}
	logrus.Infof("sucessfully retrieved collection list with cursor %s", a.AccountCollection.Name())
	return cursor, nil
}
