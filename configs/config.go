package configs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App       `mapstructure:"app"`
	Tarantool `mapstructure:"tarantool"`
	Telegram  `mapstructure:"telegram"`
	Rest      `mapstructure:"rest"`
	Template  `mapstructure:"template"`
	Features  []string `mapstructure:"features"`
}

type App struct {
	HostURL string `mapstructure:"host_url"`
	Street  `mapstructure:"street"`
	Map     `mapstructure:"map"`
}

type Street struct {
	CacheDuration time.Duration `mapstructure:"cache_duration"`
}

type Map struct {
	CacheDuration time.Duration `mapstructure:"cache_duration"`
}

type Tarantool struct {
	Servers           []string      `mapstructure:"servers"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Timeout           time.Duration `mapstructure:"timeout"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
}

type Telegram struct {
	BotToken       string `mapstructure:"bot_token"`
	Timeout        int    `mapstructure:"timeout"`
	FeedbackChatID int64  `mapstructure:"feedback_chat_id"`
}

type Rest struct {
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

type Template struct {
	Folders []string `mapstructure:"folders"`
}

func Init() (*Config, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading env variables: %w", err)
	}

	cfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_FEEDBACK_CHAT_ID"), 10, 0)
	if err != nil {
		return nil, fmt.Errorf("error convert feedback chat: %w", err)
	}
	cfg.Telegram.FeedbackChatID = chatID
	cfg.App.HostURL = os.Getenv("HOST_URL")

	return &cfg, nil
}

type configKey struct{}

func Set(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

func Get(ctx context.Context) *Config {
	contextConfig, _ := ctx.Value(configKey{}).(*Config)

	return contextConfig
}
