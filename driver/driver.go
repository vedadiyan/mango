// Copyright 2026 Pouya Vedadiyan. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package driver wraps the MongoDB Go driver with functional options and a
// client-wide BSON registry.
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
	// Aliases for driver types, so callers need not import the driver's
	// packages directly.
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

	// ClientOptions configures a Client.
	ClientOptions struct {
		*options.ClientOptions
	}

	// Client is a connection to a MongoDB deployment. It is safe for
	// concurrent use.
	Client struct {
		client   *mongo.Client
		registry *bson.Registry
	}

	// DatabaseOptions configures a Database.
	DatabaseOptions struct {
		*options.DatabaseOptionsBuilder
	}

	// Database is a handle to a MongoDB database.
	Database struct {
		database *mongo.Database
	}

	// ClientOption configures a Client.
	ClientOption func(*ClientOptions)

	// DatabaseOption configures a Database.
	DatabaseOption func(*DatabaseOptions)
)

// Connect creates a client for the deployment at the connection string url.
// Options are applied after the connection string, so they override it.
// Unless WithClientRegistry is used, the client gets a new BSON registry.
//
// Like the driver's Connect, it does not check that the deployment is
// reachable.
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

// Disconnect closes the client's connections, waiting for in-use connections
// to be returned until ctx is done.
func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Database returns a handle to the named database, configured by opts.
func (c *Client) Database(name string, opts ...DatabaseOption) *Database {
	databaseOptions := &DatabaseOptions{}
	for _, opt := range opts {
		opt(databaseOptions)
	}
	return &Database{c.client.Database(name, databaseOptions.DatabaseOptionsBuilder)}
}

// WithClientAppName sets the client's AppName option.
func WithClientAppName(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.AppName = v
	}
}

// WithClientAuth sets the client's Auth option.
func WithClientAuth(v *Credential) ClientOption {
	return func(o *ClientOptions) {
		o.Auth = v
	}
}

// WithClientAutoEncryptionOptions sets the client's AutoEncryptionOptions option.
func WithClientAutoEncryptionOptions(v *AutoEncryptionOptions) ClientOption {
	return func(o *ClientOptions) {
		o.AutoEncryptionOptions = v
	}
}

// WithClientConnectTimeout sets the client's ConnectTimeout option.
func WithClientConnectTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.ConnectTimeout = v
	}
}

// WithClientCompressors sets the client's Compressors option.
func WithClientCompressors(v []string) ClientOption {
	return func(o *ClientOptions) {
		o.Compressors = v
	}
}

// WithClientDialer sets the client's Dialer option.
func WithClientDialer(v ContextDialer) ClientOption {
	return func(o *ClientOptions) {
		o.Dialer = v
	}
}

// WithClientDirect sets the client's Direct option.
func WithClientDirect(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.Direct = v
	}
}

// WithClientDisableCertificateRevocationCheck sets the client's DisableCertificateRevocationCheck option.
func WithClientDisableCertificateRevocationCheck(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.DisableCertificateRevocationCheck = v
	}
}

// WithClientDisableOCSPEndpointCheck sets the client's DisableOCSPEndpointCheck option.
func WithClientDisableOCSPEndpointCheck(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.DisableOCSPEndpointCheck = v
	}
}

// WithClientDriverInfo sets the client's DriverInfo option.
func WithClientDriverInfo(v *DriverInfo) ClientOption {
	return func(o *ClientOptions) {
		o.DriverInfo = v
	}
}

// WithClientHeartbeatInterval sets the client's HeartbeatInterval option.
func WithClientHeartbeatInterval(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.HeartbeatInterval = v
	}
}

// WithClientHosts sets the client's Hosts option.
func WithClientHosts(v []string) ClientOption {
	return func(o *ClientOptions) {
		o.Hosts = v
	}
}

// WithClientHTTPClient sets the client's HTTPClient option.
func WithClientHTTPClient(v *http.Client) ClientOption {
	return func(o *ClientOptions) {
		o.HTTPClient = v
	}
}

// WithClientLoadBalanced sets the client's LoadBalanced option.
func WithClientLoadBalanced(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.LoadBalanced = v
	}
}

// WithClientLocalThreshold sets the client's LocalThreshold option.
func WithClientLocalThreshold(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.LocalThreshold = v
	}
}

// WithClientLoggerOptions sets the client's LoggerOptions option.
func WithClientLoggerOptions(v *LoggerOptions) ClientOption {
	return func(o *ClientOptions) {
		o.LoggerOptions = v
	}
}

// WithClientMaxConnIdleTime sets the client's MaxConnIdleTime option.
func WithClientMaxConnIdleTime(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.MaxConnIdleTime = v
	}
}

// WithClientMaxPoolSize sets the client's MaxPoolSize option.
func WithClientMaxPoolSize(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MaxPoolSize = v
	}
}

// WithClientMinPoolSize sets the client's MinPoolSize option.
func WithClientMinPoolSize(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MinPoolSize = v
	}
}

// WithClientMaxConnecting sets the client's MaxConnecting option.
func WithClientMaxConnecting(v *uint64) ClientOption {
	return func(o *ClientOptions) {
		o.MaxConnecting = v
	}
}

// WithClientPoolMonitor sets the client's PoolMonitor option.
func WithClientPoolMonitor(v *PoolMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.PoolMonitor = v
	}
}

// WithClientMonitor sets the client's Monitor option.
func WithClientMonitor(v *CommandMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.Monitor = v
	}
}

// WithClientServerMonitor sets the client's ServerMonitor option.
func WithClientServerMonitor(v *ServerMonitor) ClientOption {
	return func(o *ClientOptions) {
		o.ServerMonitor = v
	}
}

// WithClientReadConcern sets the client's ReadConcern option.
func WithClientReadConcern(v *ReadConcern) ClientOption {
	return func(o *ClientOptions) {
		o.ReadConcern = v
	}
}

// WithClientReadPreference sets the client's ReadPreference option.
func WithClientReadPreference(v *ReadPref) ClientOption {
	return func(o *ClientOptions) {
		o.ReadPreference = v
	}
}

// WithClientBSONOptions sets the client's BSONOptions option.
func WithClientBSONOptions(v *BSONOptions) ClientOption {
	return func(o *ClientOptions) {
		o.BSONOptions = v
	}
}

// WithClientRegistry sets the client's Registry option.
func WithClientRegistry(v *Registry) ClientOption {
	return func(o *ClientOptions) {
		o.Registry = v
	}
}

// WithClientReplicaSet sets the client's ReplicaSet option.
func WithClientReplicaSet(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.ReplicaSet = v
	}
}

// WithClientRetryReads sets the client's RetryReads option.
func WithClientRetryReads(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.RetryReads = v
	}
}

// WithClientRetryWrites sets the client's RetryWrites option.
func WithClientRetryWrites(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.RetryWrites = v
	}
}

// WithClientServerAPIOptions sets the client's ServerAPIOptions option.
func WithClientServerAPIOptions(v *ServerAPIOptions) ClientOption {
	return func(o *ClientOptions) {
		o.ServerAPIOptions = v
	}
}

// WithClientServerMonitoringMode sets the client's ServerMonitoringMode option.
func WithClientServerMonitoringMode(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.ServerMonitoringMode = v
	}
}

// WithClientServerSelectionTimeout sets the client's ServerSelectionTimeout option.
func WithClientServerSelectionTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.ServerSelectionTimeout = v
	}
}

// WithClientSRVMaxHosts sets the client's SRVMaxHosts option.
func WithClientSRVMaxHosts(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.SRVMaxHosts = v
	}
}

// WithClientSRVServiceName sets the client's SRVServiceName option.
func WithClientSRVServiceName(v *string) ClientOption {
	return func(o *ClientOptions) {
		o.SRVServiceName = v
	}
}

// WithClientTimeout sets the client's Timeout option.
func WithClientTimeout(v *time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.Timeout = v
	}
}

// WithClientTLSConfig sets the client's TLSConfig option.
func WithClientTLSConfig(v *tls.Config) ClientOption {
	return func(o *ClientOptions) {
		o.TLSConfig = v
	}
}

// WithClientWriteConcern sets the client's WriteConcern option.
func WithClientWriteConcern(v *WriteConcern) ClientOption {
	return func(o *ClientOptions) {
		o.WriteConcern = v
	}
}

// WithClientZlibLevel sets the client's ZlibLevel option.
func WithClientZlibLevel(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.ZlibLevel = v
	}
}

// WithClientZstdLevel sets the client's ZstdLevel option.
func WithClientZstdLevel(v *int) ClientOption {
	return func(o *ClientOptions) {
		o.ZstdLevel = v
	}
}

// WithClientMaxAdaptiveRetries sets the client's MaxAdaptiveRetries option.
func WithClientMaxAdaptiveRetries(v *uint) ClientOption {
	return func(o *ClientOptions) {
		o.MaxAdaptiveRetries = v
	}
}

// WithClientEnableOverloadRetargeting sets the client's EnableOverloadRetargeting option.
func WithClientEnableOverloadRetargeting(v *bool) ClientOption {
	return func(o *ClientOptions) {
		o.EnableOverloadRetargeting = v
	}
}

// WithDatabaseReadConcern sets the database's ReadConcern option.
func WithDatabaseReadConcern(v *ReadConcern) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetReadConcern(v)
	}
}

// WithDatabaseWriteConcern sets the database's WriteConcern option.
func WithDatabaseWriteConcern(v *WriteConcern) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetWriteConcern(v)
	}
}

// WithDatabaseReadPreference sets the database's ReadPreference option.
func WithDatabaseReadPreference(v *ReadPref) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetReadPreference(v)
	}
}

// WithDatabaseBSONOptions sets the database's BSONOptions option.
func WithDatabaseBSONOptions(v *BSONOptions) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetBSONOptions(v)
	}
}

// WithDatabaseRegistry sets the database's Registry option.
func WithDatabaseRegistry(v *Registry) DatabaseOption {
	return func(o *DatabaseOptions) {
		o.SetRegistry(v)
	}
}
