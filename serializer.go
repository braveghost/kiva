package apollo

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego/validation"
	"gopkg.in/yaml.v2"
)

type (
	configType string
	configSerializer func(buf []byte, c validation.ValidFormer) error
)

const (
	TypeJson configType = "json"
	TypeYaml configType = "yaml"
	TypeToml configType = "toml"
)

var (
	//	config serializer
	defaultSerializer = YamlSerializer
	// config Type
	defaultConfigType = TypeYaml
)

// 设置配置类型
func SetConfigType(tp configType) {
	if len(tp) != 0 {
		defaultConfigType = tp
	}
	switch defaultConfigType {
	case TypeJson:
		defaultSerializer = JsonSerializer
	case TypeToml:
		defaultSerializer = TomlSerializer
	case TypeYaml:
		defaultSerializer = YamlSerializer
	}
}

// 设置序列化器
func SetSerializer(fn configSerializer) {
	defaultSerializer = fn
}

// json 序列化
func JsonSerializer(buf []byte, c validation.ValidFormer) error {
	return json.Unmarshal(buf, &c)
}

// yaml 序列化
func YamlSerializer(buf []byte, c validation.ValidFormer) error {
	return yaml.Unmarshal(buf, c)
}

// yaml 序列化
func TomlSerializer(buf []byte, c validation.ValidFormer) error {
	return toml.Unmarshal(buf, c)
}
