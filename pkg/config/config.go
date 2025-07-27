package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string          `mapstructure:"environment"`
	Server      ServerConfig    `mapstructure:"server"`
	Database    DatabaseConfig  `mapstructure:"database"`
	Redis       RedisConfig     `mapstructure:"redis"`
	IPFS        IPFSConfig      `mapstructure:"ipfs"`
	Filecoin    FilecoinConfig  `mapstructure:"filecoin"`
	Pricing     PricingConfig   `mapstructure:"pricing"`
	Workers     WorkersConfig   `mapstructure:"workers"`
	JWT         JWTConfig       `mapstructure:"jwt"`
	RateLimit   RateLimitConfig `mapstructure:"rate_limit"`
	Logging     LoggingConfig   `mapstructure:"logging"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func (c ServerConfig) Address() string {
	return c.Host + ":" + string(rune(c.Port))
}

type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	URL       string `mapstructure:"url"`
	Namespace string `mapstructure:"namespace"`
}

func (c RedisConfig) Pool() *redis.Pool {
	return &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(c.URL)
		},
	}
}

type IPFSConfig struct {
	APIURL     string        `mapstructure:"api_url"`
	GatewayURL string        `mapstructure:"gateway_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

type FilecoinConfig struct {
	LotusAPI        string `mapstructure:"lotus_api"`
	LotusToken      string `mapstructure:"lotus_token"`
	WalletAddress   string `mapstructure:"wallet_address"`
	MinDealDuration int64  `mapstructure:"min_deal_duration"`
}

type PricingConfig struct {
	BasePricePerGBPerMonth float64 `mapstructure:"base_price_per_gb_per_month"`
	MarkupPercentage       float64 `mapstructure:"markup_percentage"`
	MinimumDealSize        int64   `mapstructure:"minimum_deal_size"`
}

type WorkersConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Enable environment variable reading
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return &config
}

func setDefaults() {
	// Environment
	viper.SetDefault("environment", "development")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.shutdown_timeout", "10s")

	// Database defaults
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	// Redis defaults
	viper.SetDefault("redis.namespace", "pinning_service")

	// IPFS defaults
	viper.SetDefault("ipfs.api_url", "http://localhost:5001")
	viper.SetDefault("ipfs.gateway_url", "http://localhost:8080")
	viper.SetDefault("ipfs.timeout", "30s")

	// Filecoin defaults
	viper.SetDefault("filecoin.lotus_api", "http://localhost:1234/rpc/v0")
	viper.SetDefault("filecoin.min_deal_duration", 518400)

	// Pricing defaults
	viper.SetDefault("pricing.base_price_per_gb_per_month", 0.001)
	viper.SetDefault("pricing.markup_percentage", 20.0)
	viper.SetDefault("pricing.minimum_deal_size", 1048576)

	// Workers defaults
	viper.SetDefault("workers.concurrency", 5)

	// JWT defaults
	viper.SetDefault("jwt.expiration", "24h")

	// Rate limiting defaults
	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.burst", 20)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}
