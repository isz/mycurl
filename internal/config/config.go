package config

import (
	"flag"
	"fmt"
	"log"
	"mycurl/internal/http"
	"os"

	"github.com/pkg/errors"
)

type Config struct {
	Verbose     bool
	Method      string
	URL         string
	ContentType string
	Body        string
}

func GetConfig() *Config {
	flag.Usage = func() {
		fmt.Printf("Usage: \ncurl [options...] <URL> <body>\n\n")
		flag.PrintDefaults()
	}
	a := Config{}
	flag.BoolVar(&a.Verbose, "v", false, "verbose")
	flag.StringVar(&a.Method, "m", "GET", "GET or POST method")
	flag.StringVar(&a.ContentType, "c", "text/plain", "content type header")

	flag.Parse()
	a.URL = flag.Arg(0)
	a.Body = flag.Arg(1)
	return &a
}

func (cfg *Config) Validate() {
	if err := cfg.validate(); err != nil {
		log.Println(err)
		flag.Usage()
		os.Exit(1)
	}
}

func (cfg *Config) validate() error {
	if cfg.URL == "" {
		return errors.New("URL is empty")
	}
	method := http.Method(cfg.Method)
	if !method.IsValid() {
		return errors.New("URL is empty")
	}
	return nil
}
