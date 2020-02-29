package test

import (
	"fmt"
	apollo "github.com/braveghost/kiva"
	"github.com/braveghost/kiva/example/properties"
	"testing"
	"time"
)

func init() {
	apollo.SetConfigType(apollo.TypeToml)
	apollo.SetPath("kiva/example/properties/local.properties")
}

func TestInitApolloSerializer(t *testing.T) {
	// 初始化配置信息

	apollo.InitApolloSerializer()

	for {
		fmt.Println(properties.App)
		time.Sleep(time.Second)
	}
}
func TestInitApolloSerializerByConf(t *testing.T) {
	// 初始化配置信息

	apollo.InitApolloSerializerByConf(apollo.GetConfig("example/properties/local.properties"))

	for {
		fmt.Println(properties.App)
		time.Sleep(time.Second)
	}
}
func TestInitApolloSerializerByConfByStr(t *testing.T) {
	// 初始化配置信息
	c := apollo.GetConfigByStr(`
{
"appId":"jd_backend",
"cluster":"default",
"namespaceNames":["test.yaml"],
"meta_addr":"111.111.111.111:8100",
"cacheDir":"/tmp"
}
`)
	apollo.InitApolloSerializerByConf(c)

	for {
		fmt.Println(properties.App)
		time.Sleep(time.Second)
	}
}
func TestInitApolloByPath(t *testing.T) {

	apollo.InitApolloByPath("/Users/miller/Documents/GoPro/src/github.com/braveghost/kiva/example/properties/local.properties")

	for {
		fmt.Println(properties.App)
		time.Sleep(time.Second)
	}
}
func TestNewApolloSerializer(t *testing.T) {

	ac := apollo.NewApolloSerializer(apollo.GetConfig("/Users/miller/Documents/GoPro/src/github.com/braveghost/kiva/example/properties/local.properties"), nil)
	ac.Start()
	for {
		fmt.Println(properties.App)
		time.Sleep(time.Second)
	}
}
