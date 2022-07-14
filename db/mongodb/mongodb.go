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
	cgClient              *mongo.Client
	cgDatabase            *mongo.Database
	ngClient              *mongo.Client
	ngDatabase            *mongo.Database
	AccountDAO            *harness.AccountDAO
	UserDAO               *harness.UserDAO
	ModuleLicenseDAO      *harness.ModuleLicenseDAO
	EnvironmentGroupNGDAO *harness.EnvironmentGroupNGDAO
}

func New(cgConf config.CGMongoDbConf, ngConf config.NGMongoDbConf) (*MongoDb, error) {
	logrus.Info("trying to connect to mongo")

	// Set up CG
	cgClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cgConf.ConnStr))
	if err != nil {
		panic(err)
	}

	// Ping mongo server to see if it's accessible. It's necessary to start GitOpsService
	err = cgClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return nil, err
	}

	logrus.Info("successfully pinged cg mongo server")
	cgDatabase := cgClient.Database(cgConf.DbName)

	accountDAO, err := harness.NewAccountDAO(cgDatabase)
	if err != nil {
		return nil, err
	}

	userDAO, err := harness.NewUserDAO(cgDatabase)
	if err != nil {
		return nil, err
	}

	// Set up NG
	ngClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(ngConf.ConnStr))
	if err != nil {
		panic(err)
	}

	// Ping mongo server to see if it's accessible. It's necessary to start GitOpsService
	err = ngClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return nil, err
	}

	logrus.Info("successfully pinged ng mongo server")
	ngDatabase := ngClient.Database(ngConf.DbName)

	moduleLicenseDAO, err := harness.NewModuleLicenseDAO(ngDatabase)
	if err != nil {
		return nil, err
	}

	envGroupDAO, err := harness.NewEnvironmentGroupNGDAO(ngDatabase)
	if err != nil {
		return nil, err
	}

	return &MongoDb{
		cgClient:              cgClient,
		cgDatabase:            cgDatabase,
		AccountDAO:            accountDAO,
		UserDAO:               userDAO,
		ModuleLicenseDAO:      moduleLicenseDAO,
		EnvironmentGroupNGDAO: envGroupDAO,
	}, nil
}
