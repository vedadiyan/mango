package driver

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type (
	Credential            = options.Credential
	AutoEncryptionOptions = options.AutoEncryptionOptions
	DriverInfo            = options.DriverInfo
	LoggerOptions         = options.LoggerOptions
	ContextDialer         = options.ContextDialer
	BSONOptions           = options.BSONOptions
	ServerAPIOptions      = options.ServerAPIOptions
	PoolMonitor           = event.PoolMonitor
	CommandMonitor        = event.CommandMonitor
	ServerMonitor         = event.ServerMonitor
	ReadConcern           = readconcern.ReadConcern
	ReadPref              = readpref.ReadPref
	Registry              = bson.Registry
	WriteConcern          = writeconcern.WriteConcern

	ClientOptions struct {
		*options.ClientOptions
	}

	Client struct {
		client   *mongo.Client
		registry *bson.Registry
	}

	DatabaseOptions struct {
		*options.DatabaseOptionsBuilder
	}

	Database struct {
		database *mongo.Database
	}

	ClientOption   func(*ClientOptions)
	DatabaseOption func(*DatabaseOptions)
)

func Connect(url string, opts ...ClientOption) (*Client, error) {
	clientOptions := &ClientOptions{}
	clientOptions.ClientOptions = &options.ClientOptions{}
	clientOptions.Registry = bson.NewRegistry()
	clientOptions.ClientOptions = clientOptions.ApplyURI(url)

	for _, opt := range opts {
		opt(clientOptions)
	}
	client, err := mongo.Connect(clientOptions.ClientOptions)
	if err != nil {
		return nil, err
	}

	return &Client{client, clientOptions.Registry}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) Database(name string, opts ...DatabaseOption) *Database {
	databaseOptions := &DatabaseOptions{}
	for _, opt := range opts {
		opt(databaseOptions)
	}
	return &Database{c.client.Database(name, databaseOptions.DatabaseOptionsBuilder)}
}

func WithClientAppName(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.AppName = v
	}
}

func WithClientAuth(v *Credential) ClientOption {
	return func(o *ClientOptions) {
		o.Auth = v
	}
}

func WithClientAutoEncryptionOptions(v *AutoEncryptionOptions) ClientOption {
	return func(o *ClientOptions) {
		o.AutoEncryptionOptions = v
	}
}

func WithClientConnectTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.ConnectTimeout = v
	}
}

func WithClientCompressors(v []string) ClientOption {
	return func(o *ClientOptions) {
		o.Compressors = v
	}
}

func WithClientDialer(v ContextDialer) ClientOption {
	return func(o *ClientOptions) {
		o.Dialer = v
	}
}

func WithClientDirect(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.Direct = v
	}
}

func WithClientDisableCertificateRevocationCheck(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.DisableCertificateRevocationCheck = v
	}
}

func WithClientDisableOCSPEndpointCheck(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.DisableOCSPEndpointCheck = v
	}
}

func WithClientDriverInfo(v *DriverInfo) ClientOption {
	return func(o *ClientOptions) {
		o.DriverInfo = v
	}
}

func WithClientHeartbeatInterval(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.HeartbeatInterval = v
	}
}

func WithClientHosts(v []string) ClientOption {
	return func(o *ClientOptions) {
		o.Hosts = v
	}
}

func WithClientHTTPClient(v *http.Client) ClientOption {
	return func(o *ClientOptions) {
		o.HTTPClient = v
	}
}

func WithClientLoadBalanced(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.LoadBalanced = v
	}
}

func WithClientLocalThreshold(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.LocalThreshold = v
	}
}

func WithClientLoggerOptions(v *LoggerOptions) ClientOption {
	return func(o *ClientOptions) {
		o.LoggerOptions = v
	}
}

func WithClientMaxConnIdleTime(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.MaxConnIdleTime = v
	}
}

func WithClientMaxPoolSize(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MaxPoolSize = v
	}
}

func WithClientMinPoolSize(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MinPoolSize = v
	}
}

func WithClientMaxConnecting(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MaxConnecting = v
	}
}

func WithClientPoolMonitor(v *PoolMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.PoolMonitor = v
	}
}

func WithClientMonitor(v *CommandMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.Monitor = v
	}
}

func WithClientServerMonitor(v *ServerMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.ServerMonitor = v
	}
}

func WithClientReadConcern(v *ReadConcern) ClientOption {
	return func(o *ClientOptions) {
		o.ReadConcern = v
	}
}

func WithClientReadPreference(v *ReadPref) ClientOption {
	return func(o *ClientOptions) {
		o.ReadPreference = v
	}
}

func WithClientBSONOptions(v *BSONOptions) ClientOption {
	return func(o *ClientOptions) {
		o.BSONOptions = v
	}
}

func WithClientRegistry(v *Registry) ClientOption {
	return func(o *ClientOptions) {
		o.Registry = v
	}
}

func WithClientReplicaSet(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.ReplicaSet = v
	}
}

func WithClientRetryReads(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.RetryReads = v
	}
}

func WithClientRetryWrites(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.RetryWrites = v
	}
}

func WithClientServerAPIOptions(v *ServerAPIOptions) ClientOption {
	return func(o *ClientOptions) {
		o.ServerAPIOptions = v
	}
}

func WithClientServerMonitoringMode(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.ServerMonitoringMode = v
	}
}

func WithClientServerSelectionTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.ServerSelectionTimeout = v
	}
}

func WithClientSRVMaxHosts(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.SRVMaxHosts = v
	}
}

func WithClientSRVServiceName(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.SRVServiceName = v
	}
}

func WithClientTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.Timeout = v
	}
}

func WithClientTLSConfig(v *tls.Config) ClientOption {
	return func(o *ClientOptions) {
		o.TLSConfig = v
	}
}

func WithClientWriteConcern(v *WriteConcern) ClientOption {
	return func(o *ClientOptions) {
		o.WriteConcern = v
	}
}

func WithClientZlibLevel(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.ZlibLevel = v
	}
}

func WithClientZstdLevel(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.ZstdLevel = v
	}
}

func WithClientMaxAdaptiveRetries(v *uint) ClientOption {
	return func(o *ClientOptions) {
		o.MaxAdaptiveRetries = v
	}
}

func WithClientEnableOverloadRetargeting(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.EnableOverloadRetargeting = v
	}
}

func WithDatabaseReadConcern(v *ReadConcern) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetReadConcern(v)
	}
}

func WithDatabaseWriteConcern(v *WriteConcern) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetWriteConcern(v)
	}
}

func WithDatabaseReadPreference(v *ReadPref) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetReadPreference(v)
	}
}

func WithDatabaseBSONOptions(v *BSONOptions) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetBSONOptions(v)
	}
}

func WithDatabaseRegistry(v *Registry) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetRegistry(v)
	}
}
