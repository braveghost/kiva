package apollo

import (
	"github.com/astaxie/beego/validation"
	"github.com/braveghost/agollo"
	logging "github.com/braveghost/joker"
	"github.com/pkg/errors"
	"os"
	"path"
	"sync"
)

type configAlarm func(string) error

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
	relativePathEnvKey = "APOLLO_RELATIVE_PATH"
	absolutePathEnvKey = "APOLLO_ABSOLUTE_PATH"

	// 默认配置文件相对路径
	relativePath string
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

func init() {
	go func() {
		for err := range errQueue {
			if alarmFunc != nil {
				innerErr := alarmFunc(err.Error())
				if innerErr != nil {
					logging.Errorw("Kiva.AlertFunc.Run.Error", "err", innerErr)
				}
			}
		}
	}()
}

func SetAlarm(fn configAlarm) {
	alarmFunc = fn
}

func SetContainer(ns string, cf validation.ValidFormer) {
	defaultConfigs[ns] = cf
}

type option func(ce *agollo.ChangeEvent) error

// 设置加载函数
func SetReloadOptions(fns map[string]option) {
	reloadOptions = fns
}

// 设置加载函数
func AddReloadOption(key string, fn option) {
	reloadOptions[key] = fn
}

// 从环境变量获取 properties 文件绝对路径
func SetAbsolutePathByEnv() {
	p := os.Getenv(absolutePathEnvKey)
	if p == "" {
		logging.Panicf("Kiva.SetCfgRelativePathByEnv.Path.Null", "err", EnvApolloConfErr)
	}
	absolutePath = p
}

// 从环境变量获取 properties 文件相对路径
func SetRelativePathByEnv() {
	p := os.Getenv(relativePathEnvKey)
	if p == "" {
		logging.Panicf("Kiva.SetCfgRelativePathByEnv.Path.Null", "err", EnvApolloConfErr)
	}
	absolutePath = p
}

// 设置 properties 文件相对路径, 如果同时设置绝对路径, 绝对路径优先
func SetRelativePath(p string) {
	relativePath = p
}

// 根据 mode 拼接配置文件相对路径
func SetRelativePathByMode(mode string) {
	relativePath = "properties/" + mode + ".properties"
}

// 设置 properties 文件绝对路径, 如果同时设置相对路径, 绝对路径优先
func SetAbsolutePath(p string) {
	absolutePath = p
}

// 初始化配置
func InitConfig() {
	if len(relativePath) == 0 && len(absolutePath) == 0 {
		logging.Errorw("Kiva.InitConfig.PropertiesPath.Error", "err", EnvApolloConfErr)
		os.Exit(1)
	}

	var pt string
	if len(absolutePath) == 0 {
		tmp, _ := os.Getwd()
		pt = path.Join(tmp, relativePath)
	} else {
		pt = absolutePath
	}
	conf, err := readConf(pt)
	if err != nil {
		logging.Errorw("Kiva.InitConfig.ReadConf.Error", "err", err)
		os.Exit(1)
	}
	err = apolloStart(conf)
	if err != nil {
		logging.Errorw("Kiva.InitConfig.ApolloStart.Error", "err", ApolloStartErr)
	}
}

func readConf(pt string) (*agollo.Conf, error) {
	logging.Infow("Kiva.ReadConf.Info", "path", pt)

	conf, err := agollo.NewConf(pt)
	if err != nil {
		logging.Errorw("Kiva.InitApolloCfg.NewConf.Error", "err", err)
		return nil, err
	}
	return conf, err
}

func InitApollo(pt string) {
	conf, err := readConf(pt)
	if err != nil {
		logging.Errorw("Kiva.InitConfig.ReadConf.Error", "err", err)
		os.Exit(1)
	}
	err = apolloStart(conf)
	if err != nil {
		logging.Errorw("Kiva.InitConfig.ApolloStart.Error", "err", ApolloStartErr)
	}
}

func apolloStart(conf *agollo.Conf) (err error) {

	configNameSpaceName = conf.NameSpaceNames

	err = agollo.StartWithConf(conf)
	if err != nil {
		logging.Errorw("Kiva.InitApolloCfg.StartWithConf.Error", "err", err, "conf", conf)
		return nil
	}

	err = reload()
	if err != nil {
		logging.Errorw("Kiva.InitApolloCfg.Load.Error", "err", err, "conf", conf)
		return err
	}
	go func() {
		for changeEvent := range agollo.WatchUpdate() {
			ns := changeEvent.Namespace
			logging.Warnw("Kiva.InitApolloCfg.apollo.WatchUpdate.Warning", ns, "配置中心变化")
			err := upload(ns)
			if err != nil {
				logging.Errorw("Kiva.InitCfg.WatchUpdate.Upload.Error", "name_space", ns, "err", err)
				errQueue <- err
			}
		}
	}()

	return
}

// 配置加载
func reload() error {
	for _, ns := range configNameSpaceName {
		err := upload(ns)
		if err != nil {
			logging.Errorw("Kiva.InitCfg.Reload.Upload.Error", "name_space", ns, "err", err)
			return err
		}
	}
	return nil
}

var lock = sync.Mutex{}

// 监听更新
func upload(ns string) error {
	lock.Lock()
	defer lock.Unlock()
	content := agollo.GetNameSpaceContent(ns, "")
	conf, ok := defaultConfigs[ns]
	if ok {
		if len(content) == 0 {
			return errors.Wrapf(GetApolloConfErr, "Namespace '%s' get", ns)
		}
		err := defaultSerializer([]byte(content), conf)
		if err != nil {
			return errors.Wrapf(SetApolloConfErr, "Namespace '%s' unmarshal [%s]", ns, err.Error())
		}
		if err := Valid(conf); err != nil {
			return errors.Wrapf(SetApolloConfErr, "Namespace '%s' valid", ns)
		}

	} else {
		return errors.Wrapf(SetApolloConfErr, "Namespace '%s'", ns)
	}

	return nil
}
