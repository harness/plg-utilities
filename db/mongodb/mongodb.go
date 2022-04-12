package mongodb

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"plg-utilities/config"
	"plg-utilities/db/harness"
)

type MongoDb struct {
	client     *mongo.Client
	database   *mongo.Database
	AccountDAO *harness.AccountDAO
	UserDAO    *harness.UserDAO
}

func New(conf config.MongoDbConf) (*MongoDb, error) {
	logrus.Info("trying to connect to mongo")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(conf.ConnStr))
	if err != nil {
		panic(err)
	}

	// Ping mongo server to see if it's accessible. It's necessary to start GitOpsService
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return nil, err
	}

	logrus.Info("successfully pinged mongo server")
	database := client.Database(conf.DbName)

	accountDAO, err := harness.NewAccountDAO(database)
	if err != nil {
		return nil, err
	}

	userDAO, err := harness.NewUserDAO(database)
	if err != nil {
		return nil, err
	}

	return &MongoDb{
		client:     client,
		database:   database,
		AccountDAO: accountDAO,
		UserDAO:    userDAO,
	}, nil
}
