package minio

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tchap/go-patricia/patricia"
	"io/ioutil"
	"net/http"
	"path"
)

// Stores system configuration, populated from CLI or test runner
type GatewayConfig struct {
	StorageDriver     StorageDriver
	BucketDriver      BucketDriver
	requestBucketChan chan BucketRequest
	dataDir           string
}

// Message for requesting a bucket
type BucketRequest struct {
	name     string
	context  Context
	callback chan Bucket
}

// Context interface for security and session information
type Context interface{}

// Bucket definition
type Bucket interface {
	GetName(Context) string
	Get(Context, string) ([]byte, error)
	Put(Context, string, []byte) error
}

// Bucket driver function, should read from a channel and respond through callback channels
type BucketDriver func(config GatewayConfig)

// Storage driver function, should read from a channel and respond through callback channels
type StorageDriver func(bucket string, input chan ObjectRequest, config GatewayConfig)

// TODO remove when building real context
type fakeContext struct{}

type GatewayGetHandler struct {
	config GatewayConfig
}

// GET requests server
func (handler GatewayGetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bucketName := vars["bucket"]
	path := vars["path"]
	context := fakeContext{}
	callback := make(chan Bucket)
	handler.config.requestBucketChan <- BucketRequest{
		name:     bucketName,
		context:  context,
		callback: callback,
	}
	bucket := <-callback
	object, err := bucket.Get(context, string(path))
	if err != nil {
		http.Error(w, err.Error(), 404)
	} else if object == nil {
		http.Error(w, errors.New("Object not found").Error(), 404)
	} else {
		fmt.Fprintf(w, string(object))
	}
}

type GatewayPutHandler struct {
	config GatewayConfig
}

func (handler GatewayPutHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	bucketName := vars["bucket"]
	path := vars["path"]
	object, _ := ioutil.ReadAll(req.Body)
	context := fakeContext{}
	callback := make(chan Bucket)
	handler.config.requestBucketChan <- BucketRequest{
		name:     bucketName,
		context:  context,
		callback: callback,
	}
	bucket := <-callback
	bucket.Put(context, path, object)
}

func RegisterGatewayHandlers(router *mux.Router, config GatewayConfig) {
	config.requestBucketChan = make(chan BucketRequest)
	go config.BucketDriver(config)
	getHandler := GatewayGetHandler{config}
	putHandler := GatewayPutHandler{config}
	router.Handle("/{bucket}/{path:.*}", getHandler).Methods("GET")
	router.Handle("/{bucket}/{path:.*}", putHandler).Methods("PUT")
}

func SynchronizedBucketDriver(config GatewayConfig) {
	buckets := make(map[string]*SynchronizedBucket)
	for request := range config.requestBucketChan {
		if buckets[request.name] == nil {
			bucketChannel := make(chan ObjectRequest)
			go config.StorageDriver(request.name, bucketChannel, config)
			buckets[request.name] = &SynchronizedBucket{
				name:    request.name,
				channel: bucketChannel,
			}
		}
		request.callback <- buckets[request.name]
	}
	for key := range buckets {
		buckets[key].closeChannel()
	}
}

type SynchronizedBucket struct {
	name    string
	channel chan ObjectRequest
	objects map[string][]byte
}

type ObjectRequest struct {
	requestType string
	path        string
	object      []byte
	callback    chan interface{}
}

func (bucket SynchronizedBucket) GetName(context Context) string {
	return bucket.name
}

func (bucket SynchronizedBucket) Get(context Context, path string) ([]byte, error) {
	callback := make(chan interface{})
	bucket.channel <- ObjectRequest{
		requestType: "GET",
		path:        path,
		callback:    callback,
	}
	response := <-callback

	switch response.(type) {
	case error:
		return nil, response.(error)
	case nil:
		return nil, errors.New("Object not found")
	case interface{}:
		return response.([]byte), nil
	default:
		return nil, errors.New("Unexpected error, service failed")
	}
}

func (bucket SynchronizedBucket) Put(context Context, path string, object []byte) error {
	callback := make(chan interface{})
	bucket.channel <- ObjectRequest{
		requestType: "PUT",
		path:        path,
		object:      object,
		callback:    callback,
	}
	switch response := <-callback; response.(type) {
	case error:
		return response.(error)
	case nil:
		return nil
	default:
		return errors.New("Unexpected error, service failed")
	}
}

func (bucket *SynchronizedBucket) closeChannel() {
	close(bucket.channel)
}

func InMemoryStorageDriver(bucket string, input chan ObjectRequest, config GatewayConfig) {
	objects := patricia.NewTrie()
	for request := range input {
		prefix := patricia.Prefix(request.path)
		switch request.requestType {
		case "GET":
			request.callback <- objects.Get(prefix)
		case "PUT":
			objects.Insert(prefix, request.object)
			request.callback <- nil
		default:
			request.callback <- errors.New("Unexpected message")
		}
	}
}

func SimpleFileStorageDriver(bucket string, input chan ObjectRequest, config GatewayConfig) {
	storage := FileStorage{
		RootDir: config.dataDir,
	}
	for request := range input {
		switch request.requestType {
		case "GET":
			objectPath := path.Join(bucket, request.path)
			object, err := storage.Get(objectPath)
			if err != nil {
				request.callback <- nil
			} else {
				request.callback <- object
			}
		case "PUT":
			objectPath := path.Join(bucket, request.path)
			storage.Put(objectPath, request.object)
			request.callback <- nil
		default:
			request.callback <- errors.New("Unexpected message")
		}
	}
}
