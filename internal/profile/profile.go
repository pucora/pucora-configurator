package profile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "https://velonetics.io/schema/v2.13/velonetics.json"

// Profile is a simplified, user-friendly configuration format.
// The generator expands it into a full velonetics.json.
type Profile struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   Meta   `yaml:"metadata"`
	Gateway    Gateway `yaml:"gateway"`
	CORS       *CORS   `yaml:"cors,omitempty"`
	Telemetry  *Telemetry `yaml:"telemetry,omitempty"`
	GRPC       *GRPC   `yaml:"grpc,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Compose    *Compose          `yaml:"compose,omitempty"`
	Routes     []Route           `yaml:"routes"`
	AsyncAgents []AsyncAgent `yaml:"async_agents,omitempty"`
}

type Meta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type Gateway struct {
	Port      int    `yaml:"port"`
	Timeout   string `yaml:"timeout,omitempty"`
	CacheTTL  string `yaml:"cache_ttl,omitempty"`
	WriteTimeout string `yaml:"write_timeout,omitempty"`
}

type CORS struct {
	Enabled        bool     `yaml:"enabled"`
	AllowOrigins   []string `yaml:"allow_origins,omitempty"`
	AllowMethods   []string `yaml:"allow_methods,omitempty"`
	AllowHeaders   []string `yaml:"allow_headers,omitempty"`
	ExposeHeaders  []string `yaml:"expose_headers,omitempty"`
	AllowCredentials bool   `yaml:"allow_credentials,omitempty"`
	MaxAge         string   `yaml:"max_age,omitempty"`
}

type Telemetry struct {
	Logging *Logging `yaml:"logging,omitempty"`
	Metrics *Metrics `yaml:"metrics,omitempty"`
	Usage   *Usage   `yaml:"usage,omitempty"`
}

type Logging struct {
	Level  string `yaml:"level,omitempty"`
	Stdout bool   `yaml:"stdout,omitempty"`
}

type Metrics struct {
	Enabled       bool   `yaml:"enabled"`
	ListenAddress string `yaml:"listen_address,omitempty"`
}

type Usage struct {
	Enabled bool `yaml:"enabled"`
}

type GRPC struct {
	Catalog []string `yaml:"catalog,omitempty"`
}

type Compose struct {
	Enabled      *bool  `yaml:"enabled,omitempty"`
	Image        string `yaml:"image,omitempty"`
	MockBackend  *bool  `yaml:"mock_backend,omitempty"`
	ExposeMetrics bool  `yaml:"expose_metrics,omitempty"`
}

type Route struct {
	Path         string       `yaml:"path"`
	Method       string       `yaml:"method"`
	Timeout      string       `yaml:"timeout,omitempty"`
	OutputEncoding string     `yaml:"output_encoding,omitempty"`
	Headers      *Headers     `yaml:"headers,omitempty"`
	QueryStrings *QueryStrings `yaml:"query_strings,omitempty"`
	Auth         *Auth        `yaml:"auth,omitempty"`
	RateLimit    *RateLimit   `yaml:"rate_limit,omitempty"`
	WebSocket    *WebSocket   `yaml:"websocket,omitempty"`
	Backend      Backend      `yaml:"backend"`
}

type Headers struct {
	Forward []string `yaml:"forward,omitempty"`
}

type QueryStrings struct {
	Forward []string `yaml:"forward,omitempty"`
}

type Auth struct {
	JWT *JWTAuth `yaml:"jwt,omitempty"`
}

type JWTAuth struct {
	Alg                  string   `yaml:"alg"`
	Audience             []string `yaml:"audience,omitempty"`
	Roles                []string `yaml:"roles,omitempty"`
	RolesKey             string   `yaml:"roles_key,omitempty"`
	Issuer               string   `yaml:"issuer,omitempty"`
	JWKURL               string   `yaml:"jwk_url"`
	DisableJWKSecurity   bool     `yaml:"disable_jwk_security,omitempty"`
	Cache                bool     `yaml:"cache,omitempty"`
}

type RateLimit struct {
	MaxRate int    `yaml:"max_rate"`
	Every   string `yaml:"every,omitempty"`
}

type WebSocket struct {
	DirectCommunication bool     `yaml:"direct_communication,omitempty"`
	MaxMessageSize      int      `yaml:"max_message_size,omitempty"`
	InputHeaders        []string `yaml:"input_headers,omitempty"`
	ConnectEvent        bool     `yaml:"connect_event,omitempty"`
	DisconnectEvent     bool     `yaml:"disconnect_event,omitempty"`
}

type Backend struct {
	Type    string `yaml:"type"` // http, grpc, kafka, nats, rabbit, websocket, graphql, soap, lambda
	Host    string `yaml:"host"`
	Path    string `yaml:"path"`
	Method  string `yaml:"method,omitempty"`
	Encoding string `yaml:"encoding,omitempty"`

	// HTTP / gRPC
	DisableHostSanitize bool `yaml:"disable_host_sanitize,omitempty"`

	// Pub/Sub
	Topic          string `yaml:"topic,omitempty"`
	Subscription   string `yaml:"subscription,omitempty"`
	ConsumerGroup  string `yaml:"consumer_group,omitempty"`

	// gRPC
	InputMapping map[string]string `yaml:"input_mapping,omitempty"`

	// GraphQL
	GraphQLType     string         `yaml:"graphql_type,omitempty"`
	QueryPath       string         `yaml:"query_path,omitempty"`
	OperationName   string         `yaml:"operation_name,omitempty"`
	GraphQLVariables map[string]any `yaml:"graphql_variables,omitempty"`

	// SOAP
	SoapTemplate string            `yaml:"soap_template,omitempty"`
	Target       string            `yaml:"target,omitempty"`
	Mapping      map[string]string `yaml:"mapping,omitempty"`
	Deny         []string          `yaml:"deny,omitempty"`

	// Kafka advanced
	KafkaCluster *KafkaCluster `yaml:"kafka_cluster,omitempty"`
}

type KafkaCluster struct {
	Brokers  []string `yaml:"brokers"`
	ClientID string   `yaml:"client_id,omitempty"`
}

type AsyncAgent struct {
	Name       string            `yaml:"name"`
	Consumer   AsyncConsumer     `yaml:"consumer"`
	Backend    AsyncBackend      `yaml:"backend"`
	Connection *AsyncConnection  `yaml:"connection,omitempty"`
	Kafka      *AsyncKafka       `yaml:"kafka,omitempty"`
}

type AsyncConnection struct {
	MaxRetries       int    `yaml:"max_retries,omitempty"`
	BackoffStrategy  string `yaml:"backoff_strategy,omitempty"`
	HealthInterval   string `yaml:"health_interval,omitempty"`
}

type AsyncConsumer struct {
	Topic   string `yaml:"topic"`
	Workers int    `yaml:"workers,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type AsyncBackend struct {
	Host string `yaml:"host"`
	Path string `yaml:"path"`
	Method string `yaml:"method,omitempty"`
}

type AsyncKafka struct {
	Brokers  []string `yaml:"brokers"`
	GroupID  string   `yaml:"group_id"`
	ClientID string   `yaml:"client_id,omitempty"`
}

func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}
	ApplyDefaults(&p)
	if err := Validate(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func UnmarshalYAML(data []byte, p *Profile) error {
	return yaml.Unmarshal(data, p)
}

func Save(path string, p *Profile) error {
	ApplyDefaults(p)
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
