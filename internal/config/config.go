package config

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type serverConfig struct {
	Addr         string `mapstructure:"addr"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	Secure       bool   `mapstructure:"secure"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type SecurityConfig struct {
	Secret string `mapstructure:"secret"`
}

type PostgresConfig struct {
	Host     string     `mapstructure:"host"`
	Port     int        `mapstructure:"port"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	DBName   string     `mapstructure:"dbname"`
	SSLMode  string     `mapstructure:"ssl_mode"`
	Pool     poolConfig `mapstructure:"pool"`
}

type poolConfig struct {
	MaxConnections        int `mapstructure:"max_connections"`
	MinConnections        int `mapstructure:"min_connections"`
	MaxConnectionLifetime int `mapstructure:"max_connection_lifetime"`
	MaxConnectionIdleTime int `mapstructure:"max_connection_idle_time"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type HTTPClientTransportConfig struct {
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout"`
	DisableCompression  bool          `mapstructure:"disable_compression"`
	TLSHandshakeTimeout time.Duration `mapstructure:"tls_handshake_timeout"`
}

type HTTPClientConfig struct {
	Timeout   time.Duration             `mapstructure:"timeout"`
	Transport HTTPClientTransportConfig `mapstructure:"transport"`
}

type LoggerConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
	FilePath string `mapstructure:"filepath"`
}

type RiotConfig struct {
	AssetsAPIBaseURL string `mapstructure:"assets_api_base_url"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type Config struct {
	Server     serverConfig     `mapstructure:"server"`
	Security   SecurityConfig   `mapstructure:"security"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
	Redis      RedisConfig      `mapstructure:"redis"`
	HTTPClient HTTPClientConfig `mapstructure:"http_client"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Riot       RiotConfig       `mapstructure:"riot"`
	CORS       CORSConfig       `mapstructure:"cors"`
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()
	v := viper.New()

	setDefault := func(v *viper.Viper) {
		v.SetDefault("server.addr", "localhost")
		v.SetDefault("server.port", 8080)
		v.SetDefault("server.mode", "debug")
		v.SetDefault("server.secure", false)
		v.SetDefault("server.read_timeout", 10)
		v.SetDefault("server.write_timeout", 10)
		v.SetDefault("server.idle_timeout", 60)

		v.SetDefault("postgres.pool.max_connections", 10)
		v.SetDefault("postgres.pool.min_connections", 1)
		v.SetDefault("postgres.pool.max_connection_lifetime", 3600)
		v.SetDefault("postgres.pool.max_connection_idle_time", 300)

		v.SetDefault("redis.host", "localhost")
		v.SetDefault("redis.port", 6379)
		v.SetDefault("redis.db", 0)
		v.SetDefault("redis.password", "")

		v.SetDefault("logger.level", "info")
		v.SetDefault("logger.encoding", "json")
		v.SetDefault("logger.filepath", "logs/app.log")

		v.SetDefault("http_client.timeout", 10)
		v.SetDefault("http_client.transport.max_idle_conns", 100)
		v.SetDefault("http_client.transport.idle_conn_timeout", 90)
		v.SetDefault("http_client.transport.disable_compression", true)
		v.SetDefault("http_client.transport.tls_handshake_timeout", 10)

		v.SetDefault("cors.allowed_origins", []string{"http://localhost:5173"})
	}

	setDefault(v)

	v.SetConfigFile("./internal/config/config.yaml")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		log.Println("Config file not found, using env/defaults only")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("server.addr", "ADDR", "APP_SERVER_ADDR")
	_ = v.BindEnv("server.port", "PORT", "APP_SERVER_PORT")
	_ = v.BindEnv("server.mode", "MODE", "APP_SERVER_MODE")
	_ = v.BindEnv("server.secure", "SECURE", "APP_SERVER_SECURE")
	_ = v.BindEnv("server.read_timeout", "READ_TIMEOUT", "APP_SERVER_READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "WRITE_TIMEOUT", "APP_SERVER_WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "IDLE_TIMEOUT", "APP_SERVER_IDLE_TIMEOUT")

	_ = v.BindEnv("security.secret", "JWT_SECRET", "APP_SECURITY_SECRET")

	_ = v.BindEnv("postgres.host", "POSTGRES_HOST", "APP_POSTGRES_HOST", "PGHOST")
	_ = v.BindEnv("postgres.port", "POSTGRES_PORT", "APP_POSTGRES_PORT", "PGPORT")
	_ = v.BindEnv("postgres.user", "POSTGRES_USER", "APP_POSTGRES_USER", "PGUSER")
	_ = v.BindEnv("postgres.password", "POSTGRES_PASSWORD", "APP_POSTGRES_PASSWORD", "PGPASSWORD")
	_ = v.BindEnv("postgres.dbname", "POSTGRES_DB", "APP_POSTGRES_DB", "PGDATABASE")

	_ = v.BindEnv("redis.host", "REDIS_HOST", "APP_REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT", "APP_REDIS_PORT")
	_ = v.BindEnv("redis.db", "REDIS_DB", "APP_REDIS_DB")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD", "APP_REDIS_PASSWORD")

	_ = v.BindEnv("logger.level", "LOG_LEVEL", "APP_LOGGER_LEVEL")
	_ = v.BindEnv("logger.encoding", "LOG_ENCODING", "APP_LOGGER_ENCODING")
	_ = v.BindEnv("logger.filename", "LOG_FILE", "APP_LOGGER_FILENAME")

	_ = v.BindEnv("riot.assets_api_base_url", "RIOT_ASSETS_API_BASE_URL", "APP_RIOT_ASSETS_API_BASE_URL")

	_ = v.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Server.Addr == "" {
		return fmt.Errorf("server addr cannot be empty")
	}
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server port cannot be zero")
	}

	if cfg.Postgres.DBName == "" {
		return fmt.Errorf("postgres dbname cannot be empty")
	}
	if cfg.Postgres.Host == "" {
		return fmt.Errorf("postgres host cannot be empty")
	}
	if cfg.Postgres.Port == 0 {
		return fmt.Errorf("postgres port cannot be zero")
	}
	if cfg.Postgres.User == "" {
		return fmt.Errorf("postgres user cannot be empty")
	}
	if cfg.Postgres.Pool.MinConnections <= 0 {
		return fmt.Errorf("postgres pool min connections cannot be zero or negative")
	}
	if cfg.Postgres.Pool.MaxConnections <= 0 {
		return fmt.Errorf("postgres pool max connections cannot be zero or negative")
	}
	if cfg.Postgres.Pool.MaxConnectionLifetime <= 0 {
		return fmt.Errorf("postgres pool max connection lifetime cannot be zero or negative")
	}
	if cfg.Postgres.Pool.MaxConnectionIdleTime <= 0 {
		return fmt.Errorf("postgres pool max connection idle time cannot be zero or negative")
	}

	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host cannot be empty")
	}
	if cfg.Redis.Port <= 0 {
		return fmt.Errorf("redis port cannot be zero or negative")
	}
	if cfg.Redis.DB < 0 {
		return fmt.Errorf("redis db cannot be negative")
	}

	if cfg.Riot.AssetsAPIBaseURL == "" {
		return fmt.Errorf("riot assets api base url cannot be empty")
	}

	return nil
}
