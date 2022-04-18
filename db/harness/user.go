package harness

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	usersCollection = "users"
)

type UserDAO struct {
	UserCollections *mongo.Collection
}

func NewUserDAO(mongoDb *mongo.Database) (*UserDAO, error) {
	dao := &UserDAO{
		UserCollections: mongoDb.Collection(usersCollection),
	}
	return dao, nil
}

//func (a *UserDAO) Create(ctx context.Context, user *core.User) (*core.User, error) {
//	_, err := a.UserCollections.InsertOne(ctx, user)
//	if err != nil {
//		logrus.WithError(err).Errorf("failed to insert document in Mongo for user %s", user.Email)
//		return nil, fmt.Errorf("failed to create application %s", user.Email)
//	}
//	return user, nil
//}
//
//func (a *UserDAO) Delete(ctx context.Context, user *core.User) (*core.User, error) {
//	//todo
//	return nil, nil
//}
//
//func (a *UserDAO) Update(ctx context.Context, user *core.User) (*core.Account, error) {
//	//todo
//	return nil, nil
//}

func (a *UserDAO) ListWithCursor(ctx context.Context) (*mongo.Cursor, error) {
	cursor, err := a.UserCollections.Find(ctx, bson.D{})
	if err != nil {
		logrus.WithError(err).Errorf("failed to retrieve documents from collection %s", usersCollection)
		return nil, err
	}
	logrus.Infof("sucessfully retrieved collection list with cursor %s", a.UserCollections.Name())
	return cursor, nil
}
