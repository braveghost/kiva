package apollo

import (
	"encoding/json"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type (
	configType       string
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

func fill(src reflect.Value, dst interface{}) error {
	dstValue := reflect.ValueOf(dst)
	dstType := reflect.TypeOf(dst).Elem()

	if dstValue.Kind() != reflect.Ptr {
		return errors.New("dst必须是point")
	}

	for i := 0; i < dstType.NumField(); i++ {
		name := dstType.Field(i).Name
		dstField := dstValue.Elem().FieldByName(name)
		if dstField.CanSet() {
			dstField.Set(src.Elem().FieldByName(name))
		}
	}
	return nil
}

// json 序列化
func JsonSerializer(buf []byte, c validation.ValidFormer) error {
	newStruct, newIter := getNewStruct(c)
	err := json.Unmarshal(buf, newIter)
	if err != nil {
		return err
	}
	return fill(newStruct, c)
}

func getNewStruct(c validation.ValidFormer) (reflect.Value, interface{}) {
	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
		t = t.Elem()
	}
	value := reflect.New(t) // 调用反射创建对象
	iter := value.Interface()
	return value, iter
}

// yaml 序列化
func YamlSerializer(buf []byte, c validation.ValidFormer) error {
	newStruct, newIter := getNewStruct(c)
	err := yaml.Unmarshal(buf, newIter)
	if err != nil {
		return err
	}
	return fill(newStruct, c)
}

// toml 序列化
func TomlSerializer(buf []byte, c validation.ValidFormer) error {
	newStruct, newIter := getNewStruct(c)
	err := toml.Unmarshal(buf, newIter)
	if err != nil {
		return err
	}
	return fill(newStruct, c)
}
