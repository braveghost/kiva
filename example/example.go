package main

import (
	"fmt"
	apollo "github.com/braveghost/kiva"
	"github.com/braveghost/kiva/example/properties"
)

func main() {
	// 初始化配置信息
	apollo.SetAbsolutePath("./example/properties/local.properties")
	apollo.SetConfigType(apollo.TypeToml)
	apollo.InitConfig()
	fmt.Println(properties.App)
}
