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
	UserLoginTTL time.Duration `toml:"user_login_ttl"` // 缓存清理天数
	DepotsNum    int           `toml:"depots_num"`     // 缓存分库数
	ModelSync    bool          `toml:"model_sync" `    // 是否同步数据库
}

func (s *_App) Valid(v *validation.Validation) {
}

func init() {
	// 实际上是 toml 格式, apollo 暂不支持 toml 编辑
	apollo.SetContainer("app.yaml", App)
}
