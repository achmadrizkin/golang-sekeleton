// Package config loads and shapes application configuration. Config is read
// from config/config.yaml (or config/<env>.yaml) via Viper, with every key
// overridable through an environment variable (uppercase, dots -> "_",
// e.g. database.host -> DATABASE_HOST). Overrides do not work reliably for
// list-typed keys (kafka.brokers, cors.allowed_origins, ...) — for
// per-environment lists use a dedicated file (config/dev.yaml,
// config/staging.yaml, config/prod.yaml) instead.
package config

// Config is the root configuration struct.
type Config struct {
	Server            ServerConfig       `mapstructure:"server"`
	Database          DatabaseConfig     `mapstructure:"database"`
	Cache             CacheConfig        `mapstructure:"cache"`
	JWT               JWTConfig          `mapstructure:"jwt"`
	CORS              CORSConfig         `mapstructure:"cors"`
	Telemetry         TelemetryConfig    `mapstructure:"telemetry"`
	Logging           LoggingConfig      `mapstructure:"logging"`
	MessagePublishers MQPublishersConfig `mapstructure:"message_publishers"`
	MessageConsumers  MQConsumersConfig  `mapstructure:"message_consumers"`
}

// ServerConfig controls transport listeners and top-level server behaviour.
type ServerConfig struct {
	Name               string `mapstructure:"name"`
	Environment        string `mapstructure:"environment"` // development|staging|production
	HTTPPort           int    `mapstructure:"http_port"`
	GRPCPort           int    `mapstructure:"grpc_port"`
	ReadTimeoutSeconds int    `mapstructure:"read_timeout_seconds"`
	ValidateJWT        bool   `mapstructure:"validate_jwt"`
	SwaggerEnabled     bool   `mapstructure:"swagger_enabled"`
}

// DatabaseConfig configures the SQL connection. Driver selects the concrete
// implementation built by internal/repository/database.NewDBReadWriter.
type DatabaseConfig struct {
	Driver                 string `mapstructure:"driver"` // mysql|postgres
	Host                   string `mapstructure:"host"`
	Port                   int    `mapstructure:"port"`
	User                   string `mapstructure:"user"`
	Password               string `mapstructure:"password"`
	Name                   string `mapstructure:"name"`
	SSLMode                string `mapstructure:"ssl_mode"` // postgres: disable|require|verify-full ...
	QueryTimeoutSeconds    int    `mapstructure:"query_timeout_seconds"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

// CacheConfig configures the primary (read/write) Redis connection and an
// optional read-only replica.
type CacheConfig struct {
	Enabled  bool             `mapstructure:"enabled"`
	Mode     string           `mapstructure:"mode"` // single|cluster
	Addrs    []string         `mapstructure:"addrs"`
	Password string           `mapstructure:"password"`
	DB       int              `mapstructure:"db"`
	Replica  *CacheReplicaCfg `mapstructure:"replica"`
}

// CacheReplicaCfg configures an optional read-only Redis replica. When set,
// Repository.ReadCache() serves reads from it instead of the master.
type CacheReplicaCfg struct {
	Mode     string   `mapstructure:"mode"`
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	DB       int      `mapstructure:"db"`
}

// JWTConfig configures Bearer-token validation.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

// CORSConfig configures pkg/middleware.CORSMiddleware.
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// TelemetryConfig configures pkg/telemetry.Init.
type TelemetryConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	SampleRatio  float64 `mapstructure:"sample_ratio"`
}

// LoggingConfig configures pkg/logger and the request-logging middleware.
type LoggingConfig struct {
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	ExcludedLogPaths []string `mapstructure:"excluded_log_paths"`
}
