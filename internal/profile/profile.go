package profile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "https://pucora.io/schema/v2.13/pucora.json"

// Profile is a simplified, user-friendly configuration format.
// The generator expands it into a full pucora.json.
type Profile struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Metadata   Meta   `yaml:"metadata" json:"metadata"`
	Gateway    Gateway `yaml:"gateway" json:"gateway"`
	CORS       *CORS   `yaml:"cors,omitempty" json:"cors,omitempty"`
	Telemetry  *Telemetry `yaml:"telemetry,omitempty" json:"telemetry,omitempty"`
	GRPC       *GRPC   `yaml:"grpc,omitempty" json:"grpc,omitempty"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Compose    *Compose          `yaml:"compose,omitempty" json:"compose,omitempty"`
	Routes     []Route           `yaml:"routes" json:"routes"`
	AsyncAgents []AsyncAgent `yaml:"async_agents,omitempty" json:"async_agents,omitempty"`
}

type Meta struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type Gateway struct {
	Port      int    `yaml:"port" json:"port"`
	Timeout   string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	CacheTTL  string `yaml:"cache_ttl,omitempty" json:"cache_ttl,omitempty"`
	WriteTimeout string `yaml:"write_timeout,omitempty" json:"write_timeout,omitempty"`
}

type CORS struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	AllowOrigins   []string `yaml:"allow_origins,omitempty" json:"allow_origins,omitempty"`
	AllowMethods   []string `yaml:"allow_methods,omitempty" json:"allow_methods,omitempty"`
	AllowHeaders   []string `yaml:"allow_headers,omitempty" json:"allow_headers,omitempty"`
	ExposeHeaders  []string `yaml:"expose_headers,omitempty" json:"expose_headers,omitempty"`
	AllowCredentials bool   `yaml:"allow_credentials,omitempty" json:"allow_credentials,omitempty"`
	MaxAge         string   `yaml:"max_age,omitempty" json:"max_age,omitempty"`
}

type Telemetry struct {
	Logging *Logging `yaml:"logging,omitempty" json:"logging,omitempty"`
	Metrics *Metrics `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Usage   *Usage   `yaml:"usage,omitempty" json:"usage,omitempty"`
}

type Logging struct {
	Level  string `yaml:"level,omitempty" json:"level,omitempty"`
	Stdout bool   `yaml:"stdout,omitempty" json:"stdout,omitempty"`
}

type Metrics struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	ListenAddress string `yaml:"listen_address,omitempty" json:"listen_address,omitempty"`
}

type Usage struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type GRPC struct {
	Catalog []string `yaml:"catalog,omitempty" json:"catalog,omitempty"`
}

type Compose struct {
	Enabled      *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Image        string `yaml:"image,omitempty" json:"image,omitempty"`
	MockBackend  *bool  `yaml:"mock_backend,omitempty" json:"mock_backend,omitempty"`
	ExposeMetrics bool  `yaml:"expose_metrics,omitempty" json:"expose_metrics,omitempty"`
}

type Route struct {
	Path         string       `yaml:"path" json:"path"`
	Method       string       `yaml:"method" json:"method"`
	Timeout      string       `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	OutputEncoding string     `yaml:"output_encoding,omitempty" json:"output_encoding,omitempty"`
	Headers      *Headers     `yaml:"headers,omitempty" json:"headers,omitempty"`
	QueryStrings *QueryStrings `yaml:"query_strings,omitempty" json:"query_strings,omitempty"`
	Auth         *Auth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	RateLimit    *RateLimit   `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
	WebSocket    *WebSocket   `yaml:"websocket,omitempty" json:"websocket,omitempty"`
	Backend      Backend      `yaml:"backend" json:"backend"`
}

type Headers struct {
	Forward []string `yaml:"forward,omitempty" json:"forward,omitempty"`
}

type QueryStrings struct {
	Forward []string `yaml:"forward,omitempty" json:"forward,omitempty"`
}

type Auth struct {
	JWT *JWTAuth `yaml:"jwt,omitempty" json:"jwt,omitempty"`
}

type JWTAuth struct {
	Alg                  string   `yaml:"alg" json:"alg"`
	Audience             []string `yaml:"audience,omitempty" json:"audience,omitempty"`
	Roles                []string `yaml:"roles,omitempty" json:"roles,omitempty"`
	RolesKey             string   `yaml:"roles_key,omitempty" json:"roles_key,omitempty"`
	Issuer               string   `yaml:"issuer,omitempty" json:"issuer,omitempty"`
	JWKURL               string   `yaml:"jwk_url" json:"jwk_url"`
	DisableJWKSecurity   bool     `yaml:"disable_jwk_security,omitempty" json:"disable_jwk_security,omitempty"`
	Cache                bool     `yaml:"cache,omitempty" json:"cache,omitempty"`
}

type RateLimit struct {
	MaxRate int    `yaml:"max_rate" json:"max_rate"`
	Every   string `yaml:"every,omitempty" json:"every,omitempty"`
}

type WebSocket struct {
	DirectCommunication bool     `yaml:"direct_communication,omitempty" json:"direct_communication,omitempty"`
	MaxMessageSize      int      `yaml:"max_message_size,omitempty" json:"max_message_size,omitempty"`
	InputHeaders        []string `yaml:"input_headers,omitempty" json:"input_headers,omitempty"`
	ConnectEvent        bool     `yaml:"connect_event,omitempty" json:"connect_event,omitempty"`
	DisconnectEvent     bool     `yaml:"disconnect_event,omitempty" json:"disconnect_event,omitempty"`
}

type Backend struct {
	Type    string `yaml:"type" json:"type"`
	Host    string `yaml:"host" json:"host"`
	Path    string `yaml:"path" json:"path"`
	Method  string `yaml:"method,omitempty" json:"method,omitempty"`
	Encoding string `yaml:"encoding,omitempty" json:"encoding,omitempty"`

	DisableHostSanitize bool `yaml:"disable_host_sanitize,omitempty" json:"disable_host_sanitize,omitempty"`

	Topic          string `yaml:"topic,omitempty" json:"topic,omitempty"`
	Subscription   string `yaml:"subscription,omitempty" json:"subscription,omitempty"`
	ConsumerGroup  string `yaml:"consumer_group,omitempty" json:"consumer_group,omitempty"`

	InputMapping map[string]string `yaml:"input_mapping,omitempty" json:"input_mapping,omitempty"`

	GraphQLType     string         `yaml:"graphql_type,omitempty" json:"graphql_type,omitempty"`
	QueryPath       string         `yaml:"query_path,omitempty" json:"query_path,omitempty"`
	OperationName   string         `yaml:"operation_name,omitempty" json:"operation_name,omitempty"`
	GraphQLVariables map[string]any `yaml:"graphql_variables,omitempty" json:"graphql_variables,omitempty"`

	SoapTemplate string            `yaml:"soap_template,omitempty" json:"soap_template,omitempty"`
	Target       string            `yaml:"target,omitempty" json:"target,omitempty"`
	Mapping      map[string]string `yaml:"mapping,omitempty" json:"mapping,omitempty"`
	Deny         []string          `yaml:"deny,omitempty" json:"deny,omitempty"`

	KafkaCluster *KafkaCluster `yaml:"kafka_cluster,omitempty" json:"kafka_cluster,omitempty"`
}

type KafkaCluster struct {
	Brokers  []string `yaml:"brokers" json:"brokers"`
	ClientID string   `yaml:"client_id,omitempty" json:"client_id,omitempty"`
}

type AsyncAgent struct {
	Name       string            `yaml:"name" json:"name"`
	Consumer   AsyncConsumer     `yaml:"consumer" json:"consumer"`
	Backend    AsyncBackend      `yaml:"backend" json:"backend"`
	Connection *AsyncConnection  `yaml:"connection,omitempty" json:"connection,omitempty"`
	Kafka      *AsyncKafka       `yaml:"kafka,omitempty" json:"kafka,omitempty"`
}

type AsyncConnection struct {
	MaxRetries       int    `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
	BackoffStrategy  string `yaml:"backoff_strategy,omitempty" json:"backoff_strategy,omitempty"`
	HealthInterval   string `yaml:"health_interval,omitempty" json:"health_interval,omitempty"`
}

type AsyncConsumer struct {
	Topic   string `yaml:"topic" json:"topic"`
	Workers int    `yaml:"workers,omitempty" json:"workers,omitempty"`
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type AsyncBackend struct {
	Host string `yaml:"host" json:"host"`
	Path string `yaml:"path" json:"path"`
	Method string `yaml:"method,omitempty" json:"method,omitempty"`
}

type AsyncKafka struct {
	Brokers  []string `yaml:"brokers" json:"brokers"`
	GroupID  string   `yaml:"group_id" json:"group_id"`
	ClientID string   `yaml:"client_id,omitempty" json:"client_id,omitempty"`
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
	data, err := MarshalYAML(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// MarshalYAML serializes a profile to YAML bytes.
func MarshalYAML(p *Profile) ([]byte, error) {
	ApplyDefaults(p)
	data, err := yaml.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal profile: %w", err)
	}
	return data, nil
}
