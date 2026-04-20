package config

import (
	"strings"

	"github.com/labstack/echo/v4"
)

type ConfigurationStruct struct {
	BaseConfig       BaseConfig
	ReportConfig     ReportConfig
	DatabaseConfig   DbConfig
	BBBServers       []BbbServer
	ScaleLiteServers []ScaleliteServer
}

func (c *ConfigurationStruct) GetBBBServer(hostname string) *BbbServer {
	for i, bbbServer := range c.BBBServers {
		if bbbServer.Hostname == hostname {
			return &c.BBBServers[i]
		}
	}
	return nil
}

func (c *ConfigurationStruct) FindBBBServers(name string) (matches []BbbServer) {
	matches = make([]BbbServer, 0)
	for _, bbbServer := range c.BBBServers {
		if strings.Contains(name, bbbServer.Hostname) || name == "" {
			matches = append(matches, bbbServer)
		}
	}
	return matches
}

func (c *ConfigurationStruct) FindScaleliteServers(name string) (matches []ScaleliteServer) {
	matches = make([]ScaleliteServer, 0)
	for _, scaleLite := range c.ScaleLiteServers {
		if strings.Contains(name, scaleLite.Hostname) || name == "" {
			matches = append(matches, scaleLite)
		}
	}
	return matches
}

func (c *ConfigurationStruct) GetScaleliteServer(hostname string) *ScaleliteServer {
	for i, scaleLite := range c.ScaleLiteServers {
		if scaleLite.Hostname == hostname {
			return &c.ScaleLiteServers[i]
		}
	}
	return nil
}

type BaseConfig struct {
	Host               string `toml:"HOST" env:"HOST" env-default:"0.0.0.0" env-description:"Host to bind bbbstatus to"`
	Port               string `toml:"PORT" env:"PORT" env-default:"8080" env-description:"Port to bind bbbstatus to"`
	ServeStaticContent bool   `toml:"SERVE_STATIC_CONTENT" env:"SERVE_STATIC_CONTENT" env-default:"true" env-description:"Serve static content using the bbbstatus binary"`
	TrustedProxies     string `toml:"TRUSTED_PROXIES" env:"TRUSTED_PROXIES" env-default:"" env-description:"Trusted proxies if behind a reverse proxy (recommended) bbbstatus to"`
	ServerLang         string `toml:"SERVER_LANG" env:"SERVER_LANG" env-default:"en" env-description:"The Server language is a fallback language"`
	ClearQueue         bool   `toml:"CLEAR_QUEUE" env:"CLEAR_QUEUE" env-default:"false" env-description:"Clears the queue of bbb-webhook events, all of them will be dropped"`
}

type ReportConfig struct {
	CsvStructure string `toml:"CSV_STRUCTURE" env:"CSV_STRUCTURE" env-default:"time,user,action,text representation" env-description:"CSV order of elements"`
}

type DbConfig struct {
	DatabaseConnectionString string `toml:"DB_CONNECTION_STRING" env:"DB_CONNECTION_STRING" env-description:"Database (PostgreSQL) connection string"`
}

type BbbServer struct {
	Hostname     string `toml:"HOSTNAME" env-required:"true"`      // required
	ApiPort      string `toml:"API_PORT"`                          // required only if not 443
	SharedSecret string `toml:"SHARED_SECRET" env-required:"true"` // required only if API usage is needed
	APITimeout   int    `toml:"API_TIMEOUT"`                       // defaults to 2 (seconds)
	FriendlyName string `toml:"FRIENDLY_NAME"`                     // not required
}

type ScaleliteServer struct {
	Hostname     string `toml:"HOSTNAME"`
	ApiPort      string `toml:"API_PORT"`
	SharedSecret string `toml:"SHARED_SECRET"`
	APITimeout   int    `toml:"API_TIMEOUT"`
	FriendlyName string `toml:"FRIENDLY_NAME"`
}

type CustomContext struct {
	echo.Context
	Config *ConfigurationStruct
}

func ConfigMiddleware(cfg *ConfigurationStruct) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c, cfg}
			return next(cc)
		}
	}
}
