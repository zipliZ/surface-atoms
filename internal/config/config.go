package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Simulating Simulating `json:"simulating"`
	Constants  Constants  `json:"consts"`
}

type Simulating struct {
	// процент шагов для записи в эксель
	LogPercent float64 `json:"logPercent"`
	MatrixLenX int     `json:"matrixLenX"`
	MatrixLenY int     `json:"matrixLenY"`
	// Тестовый параметр для проверки увеличения времени симуляции
	AllowTimeProgressInIdle bool `json:"allowTimeProgressInIdle"`
}

type Constants struct {
	Mass      float64 `json:"mass"`
	Edes      float64 `json:"edes"`
	Edif      float64 `json:"edif"`
	Vdes      float64 `json:"vdes"`
	Vdif      float64 `json:"vdif"`
	Er        float64 `json:"er"`
	Erlh      float64 `json:"erlh"`
	FDensity  float64 `json:"fDensity"`
	Fi        float64 `json:"fi"`
	SDensity  float64 `json:"sDensity"`
	AgDensity float64 `json:"agDensity"`
}

func New() (Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider("./config.yaml"), yaml.Parser()); err != nil {
		return Config{}, err
	}

	var configStruct Config
	if err := k.UnmarshalWithConf("", &configStruct, koanf.UnmarshalConf{Tag: "json"}); err != nil {
		return Config{}, err
	}

	return configStruct, nil
}
