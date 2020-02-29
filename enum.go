package apollo

import (
	"github.com/astaxie/beego/validation"
	"github.com/pkg/errors"
)

var (
	EnvApolloConfErr = errors.New("Environment [APOLLOCONFIG] variable error")
	ApolloStartErr   = errors.New("Init apollo conf error")
	GetApolloConfErr = errors.New("Get apollo conf error")
	SetApolloConfErr = errors.New("Set apollo conf error")
)

var (
	// 指定命名空间
	configNameSpaceName []string

	// 环境变量
	absolutePathEnvKey = "APOLLO_ABSOLUTE_PATH"

	// 绝对路径, 优先找绝对路径
	absolutePath string

	// reload function
	reloadOptions map[string]option

	defaultConfigs = map[string]validation.ValidFormer{}

	// error channel
	errQueue = make(chan error)
	// error alarm
	alarmFunc configAlarm
)
