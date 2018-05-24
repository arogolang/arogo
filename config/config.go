package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/arogolang/arogo/errlog"
)

var (
	// File specifies a file from which to read the config
	// If empty, config will be read from the environment
	File string

	instance      *Config
	instantiation = sync.Once{}
)

// Config holds the global application configuration.
type Config struct {
	Debug bool `json:"debug"`

	DisableWebsocket bool   `default:"true" json:"nowebsocket"`
	DisableStratum   bool   `default:"true" json:"nostratum"`
	DisableWeb       bool   `default:"false" json:"noweb"`
	PoolStartumAddr  string `default:":8888" json:"stratumaddr"`
	PoolWebAddr      string `default:":8080" json:"webaddr"`

	ShareValidation int `json:"validateshares" default:"2"`

	// LogFile and DiscardLog are mutually exclusive - logfile will be used if present
	LogFile    string `json:"log"`
	DiscardLog bool   `json:"nolog"`
}

// IsMissingConfig returns true if the the error has to do with missing required configs
func IsMissingConfig(err error) bool {
	return strings.Contains(err.Error(), "required key")
}

// only for config from file
func setDefaults(c *Config) error {
	// TODO cleanup?
	val := reflect.ValueOf(c)
	refType := reflect.TypeOf(c)
	for i := 0; i < val.Elem().NumField(); i++ {
		field := val.Elem().Field(i)
		fieldType := field.Type()
		defaultValue := refType.Elem().Field(i).Tag.Get("default")
		if defaultValue != "" {
			valueType := fieldType.Kind()
			switch valueType {
			case reflect.String:
				if field.String() == "" && field.CanSet() {
					field.SetString(defaultValue)
				}
			case reflect.Int:
				intVal, err := strconv.Atoi(defaultValue)
				if err != nil {
					return fmt.Errorf("unable to convert default value to int: %v - err: %s", defaultValue, err)
				}
				if field.Int() == 0 && field.CanSet() {
					field.SetInt(int64(intVal))
				}
			case reflect.Bool:
				if field.CanSet() {
					v, err := strconv.ParseBool(defaultValue)
					if err != nil {
						return fmt.Errorf("unable to parse bool value for: %v - err: %s"+defaultValue, err)
					}
					field.SetBool(v)
				}
			default:
				errlog.Error("Unexpected type found in config.  Skipping: ", field)
			}
		}
	}

	return nil
}

// only for config from file
func validate(c *Config) error {
	val := reflect.ValueOf(c)
	refType := reflect.TypeOf(c)
	for i := 0; i < val.Elem().NumField(); i++ {
		field := val.Elem().Field(i)

		// required fields are all strings
		if _, ok := refType.Elem().Field(i).Tag.Lookup("required"); ok && field.String() == "" {
			return fmt.Errorf("required key %s missing value", refType.Elem().Field(i).Name)
		}
	}

	return nil
}

func configFromFile(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		errlog.Error("Failed to read config file.", err.Error())
		return err
	}

	cfg := Config{}
	err = setDefaults(&cfg)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		errlog.Error("Failed to parse JSON.", err.Error())
		return err
	}
	err = validate(&cfg)
	if err != nil {
		return err
	}

	instance = &cfg
	return nil
}

// Get returns the global configuration singleton.
func Get() *Config {
	var err error
	instantiation.Do(func() {
		if File != "" {
			var f *os.File
			f, err = os.Open(File)
			if err != nil {
				errlog.Fatal("open config file failed: ", err)
				return
			}
			defer f.Close()
			err = configFromFile(f)
		} else {
			errlog.Error("config file not set")
		}
	})

	if err != nil {
		errlog.Fatal("Unable to load config: ", err)
	}
	return instance
}
