package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

import "sync"

import "net/http"

func main() {
	redisServiceBroker := &redisServiceBroker{}
	/*
	   Initialize lager.logger to pass to broker
	*/

	logger := lager.NewLogger("my-app")
	var writer *copyWriter
	fmt.Println("we made it this far")
	writer = NewCopyWriter()
	logger.RegisterSink(lager.NewWriterSink(writer, lager.INFO))
	logger.Info("lager.logger has been initialized")
	/*
	   Setup credentials
	*/
	credentials := brokerapi.BrokerCredentials{
		Username: "username",
		Password: "password",
	}
	brokerAPI := brokerapi.New(redisServiceBroker, logger, credentials)
	logger.Info("brokerApi intialized")
	http.Handle("/", brokerAPI)
	cfPort := os.Getenv("PORT")
	http.ListenAndServe(":"+cfPort, nil)
}

// copyWriter is an INTENTIONALLY UNSAFE writer. Use it to test code that
// should be handling thread safety.
type copyWriter struct {
	contents []byte
	lock     *sync.RWMutex
}

func NewCopyWriter() *copyWriter {
	return &copyWriter{
		contents: []byte{},
		lock:     new(sync.RWMutex),
	}
}

// no, we really mean RLock on write.
func (writer *copyWriter) Write(p []byte) (n int, err error) {
	writer.lock.RLock()
	defer writer.lock.RUnlock()

	writer.contents = append(writer.contents, p...)
	return len(p), nil
}

func (writer *copyWriter) Copy() []byte {
	writer.lock.Lock()
	defer writer.lock.Unlock()

	contents := make([]byte, len(writer.contents))
	copy(contents, writer.contents)
	return contents
}

type redisServiceBroker struct {
	LastOperationState       brokerapi.LastOperationState
	LastOperationDescription string
}

func (*redisServiceBroker) Services() []brokerapi.Service {
	// Return a []brokerapi.Service here, describing your service(s) and plan(s)
	return []brokerapi.Service{
		brokerapi.Service{
			ID:            "AWS/us-east-1",
			Name:          "onboardingRedis",
			Description:   "Redis service for an application that wants a redis DB",
			Bindable:      true,
			PlanUpdatable: true,
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					ID:          "1",
					Name:        "Standard 30",
					Description: "Default Redis plan",
				},
			},
			Tags: []string{
				"pivotal",
				"redis",
			},
		},
	}
}

func (*redisServiceBroker) Provision(
	instanceID string,
	details brokerapi.ProvisionDetails,
	asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {
	// Provision a new instance here. If async is allowed, the broker can still
	// chose to provision the instance synchronously.

	//TODO Removed hardcoded dashboard URL
	return brokerapi.ProvisionedServiceSpec{DashboardURL: "dashboardURL", IsAsync: false, OperationData: "blah"}, nil
}

func (rsb *redisServiceBroker) LastOperation(instanceID, operationData string) (brokerapi.LastOperation, error) {
	// If the broker provisions asynchronously, the Cloud Controller will poll this endpoint
	// for the status of the provisioning operation.
	// This also applies to deprovisioning (work in progress).
	return brokerapi.LastOperation{State: rsb.LastOperationState, Description: rsb.LastOperationDescription}, nil
}

func (rsb *redisServiceBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	// Deprovision a new instance here. If async is allowed, the broker can still
	// chose to deprovision the instance synchronously, hence the first return value.
	return brokerapi.DeprovisionServiceSpec{IsAsync: false}, nil
}

func (*redisServiceBroker) Bind(instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	// Bind to instances here
	// Return a binding which contains a credentials object that can be marshalled to JSON,
	// and (optionally) a syslog drain URL.
	return brokerapi.Binding{
		Credentials: RedisCredentials{
			Host:     "pub-redis-11035.us-east-1-3.6.ec2.redislabs.com",
			Port:     11035,
			DbName:   "onboardingRedis",
			Password: "mjh",
		},
	}, nil
}

type RedisCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DbName   string `json:"dbname"`
	Password string `json:"password"`
}

func (*redisServiceBroker) Unbind(instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	// Unbind from instances here
	return nil
}

func (*redisServiceBroker) Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	// Update instance here
	return brokerapi.UpdateServiceSpec{IsAsync: false}, nil
}
