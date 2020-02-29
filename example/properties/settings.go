package properties

import (
	"github.com/astaxie/beego/validation"
	apollo "github.com/braveghost/kiva"
	"time"
)

var (
	App = &_App{}
)

type _App struct {
	A time.Duration `toml:"a"`
	B int           `toml:"b"`
	C bool          `toml:"c" `
}

func (s *_App) Valid(v *validation.Validation) {
}

func init() {
	// 实际上是 toml 格式, apollo 暂不支持 toml 编辑
	apollo.SetContainer("test.yaml", App)
}
