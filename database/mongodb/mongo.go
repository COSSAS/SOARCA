package mongodb

import (
	"context"
	"errors"
	"reflect"
	"time"

	"soarca/database/projections"
	cacao "soarca/models/cacao"

	"go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"
)

const writeErrorDuplicationCode = 11000

var (
	cacaoPlayBookRepo *mongoCollection[cacao.Playbook]
	mongoclient       *mongo.Client
)

type dbtypes interface {
	cacao.Playbook // | for other supported types
}

type mongoCollection[T dbtypes] struct {
	Collection     *mongo.Collection
	collectionname string
}

type mongoFindOptions struct {
	findOptions *options.FindOptions
}

// type additionalFindOptions func(*mongoFindOptions)

func DefaultLimitOpts() mongoFindOptions {
	return mongoFindOptions{
		findOptions: options.Find().SetSkip(0).SetLimit(100),
	}
}

func (mongoOpts mongoFindOptions) GetIds() interface{} {
	return func(lo *mongoFindOptions) {
		lo.findOptions.SetProjection(projections.Id.GetProjection())
	}
}

func (mongoOpts mongoFindOptions) GetProjectionByType(interface{}) interface{} {
	return func(lo *mongoFindOptions) {
		lo.findOptions.SetProjection(projections.Meta.GetProjection())
	}
}

func GetCacaoRepo() *mongoCollection[cacao.Playbook] {
	return cacaoPlayBookRepo
}

// func GetMongoClient() *mongodbClient {
// 	return mongoclient
// }

func SetupMongodb(uri string, username string, password string) error {
	log.Trace("Calling SetupMongodb() to start setting up the mongodb database implementation")
	err := InitMongoClient(uri, username, password)
	if err != nil {
		log.Error("Failed to setup MongoClient, error msg: ", err.Error())
		return err
	}

	if mongoclient == nil {
		const error_msg = "Mongoclient is not set properly"
		log.Error(error_msg)
		return errors.New(error_msg)
	}

	cacaoPlayBookRepo, err = NewMongoCollection[cacao.Playbook](mongoclient, "soarca", "cacoa_playbook_collection")
	return err
}

// helper function to poperly obtain whether object is already in the database store
func isDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == writeErrorDuplicationCode { // duplication code
				return true
			}
		}
	}
	return false
}

func (mongocollection *mongoCollection[T]) Read(id string) (any, error) {
	var collection T
	context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mongocollection.Collection.FindOne(context, bson.M{"_id": id}).Decode(&collection)
	if err != nil {
		log.Debug("Error in FindOne operation: ", err)
		return nil, err
	}
	return collection, err
}

func (mongocollection *mongoCollection[T]) Find(query map[string]string, findOps ...interface{}) ([]interface{}, error) {
	opts := DefaultLimitOpts()
	for _, fn := range findOps {
		optionFunction := fn.(func(*mongoFindOptions))
		optionFunction(&opts)
	}

	bsonQuery := bson.D{}

	for key, value := range query {
		bsonQuery = append(bsonQuery, bson.E{Key: key, Value: value})
	}

	collection := make([]T, 0)

	// needs to do some hacking to get the required generic output format as any does not work.
	context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cursors, err := mongocollection.Collection.Find(context, bsonQuery, opts.findOptions)
	if err != nil {
		log.Debug("Can not find object for given findOptions: ", err)
		return nil, err
	}

	err = cursors.All(context, &collection)

	if err != nil {
		log.Debug("Cursor for objects error: ", err)
		return nil, err
	}
	interfaceCollection := make([]interface{}, len(collection))
	for i, v := range collection {
		interfaceCollection[i] = v
	}

	return interfaceCollection, nil
}

func (mongocollection *mongoCollection[T]) Create(data interface{}) error {
	var t T
	log.Trace("Attempting to insert object in collection: ", mongocollection.collectionname)
	log.Trace(data)
	input_type := reflect.TypeOf(data) // unpacking because input type of pointer type
	comparison_type := reflect.TypeOf(t)

	if input_type != comparison_type { // checks if the input type provided is accordance with collection type
		log.Error("Failed as input object does not match required collection type")
		return errors.New("Input type for Create function is not in accordance with required type")
	}
	// ctx := context.Background()
	context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := mongocollection.Collection.InsertOne(context, data)

	if isDuplicate(err) {
		log.Debug("Detected duplication in collection: ", mongocollection.collectionname, context)
		return errors.New("duplicate")
	}

	return err
}

func (mongocollection *mongoCollection[T]) Update(id string, data interface{}) error {
	var collectionFind T
	context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mongocollection.Collection.FindOne(context, bson.M{"_id": id}).Decode(&collectionFind)
	if err != nil {
		log.Debug(err)
		return err
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": data}

	_, err = mongocollection.Collection.UpdateOne(context, filter, update)
	if err != nil {
		log.Debug(err)
	}
	return err
}

func (mongocollection *mongoCollection[T]) Delete(id string) error {
	log.Trace("Trying to delete object type: ", mongocollection.collectionname, "with id: ", id)
	context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := mongocollection.Collection.DeleteOne(context, bson.M{"_id": id})
	if err != nil {
		log.Debug("Error in deleting document: ", err.Error())
	}
	return err
}

func NewMongoCollection[T dbtypes](mongo *mongo.Client, dbName string, colName string) (*mongoCollection[T], error) {
	log.Trace("Setting a new Mongo Collection name: ", colName)

	if mongo == nil {
		log.Error("nil pointer for mongoClient")
		return nil, errors.New("nil pointer for mongoClient")
	}

	if colName == "" {
		log.Error("column name for NewMongoCollection not valid, because empty")
		return nil, errors.New("database name for NewMongoCollection not valid, because empty")
	}

	if dbName == "" {
		log.Error("database name for NewMongoCollection not valid, because empty")
		return nil, errors.New("database name for NewMongoCollection not valid, because empty")
	}

	collection := mongo.Database(dbName).Collection(colName)
	return &mongoCollection[T]{Collection: collection, collectionname: colName}, nil
}

func InitMongoClient(mongo_uri string, username string, password string) error {
	log.Trace("Trying to setup new MongoClient")
	var err error
	if mongo_uri == "" {
		log.Error("mongo uri not valid, because empty")
		return errors.New("database name for NewMongoCollection not valid, because empty")
	}

	if username == "" || password == "" {
		log.Error("you must set your 'username' or 'password'")
		return errors.New("Username or password not correctly set")
	}

	clientOpts := options.Client().ApplyURI(mongo_uri).SetAuth(options.Credential{
		Username: username,
		Password: password,
	},
	)

	newContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoclient, err = mongo.Connect(newContext, clientOpts)
	const error_msg = "can't verify a connection"
	if err != nil {
		log.Error(error_msg)
		return err
	}

	err = mongoclient.Ping(context.Background(), nil)

	if err != nil {
		log.Error(error_msg)
		return err
	}

	log.Trace("Looks like we succesfully setup a Mongodb client!")
	return nil
}

func CloseMongoDB() error {
	err := mongoclient.Disconnect(context.Background())
	if err != nil {
		log.Fatal("Failed to close mongoDB connection", err.Error())
		return err
	}
	return nil
}
