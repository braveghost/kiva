package apollo

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/astaxie/beego/validation"
	logging "github.com/braveghost/joker"
	"github.com/philchia/agollo/v3"
	"github.com/pkg/errors"
)

type configAlarm func(string) error

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

// 设置告警
func SetAlarm(fn configAlarm) {
	alarmFunc = fn
}

// 配置容器设置
func SetContainer(ns string, cf validation.ValidFormer) {
	defaultConfigs[ns] = cf
}

type (
	option      func(ce *agollo.ChangeEvent) error
	watchOption func(namespace, data string)
)

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

// 设置 properties 文件绝对路径
func SetPath(p string) {
	absolutePath = p
}

// 通过 properties 初始化配置 自动序列化
func InitApolloSerializer() {
	if len(absolutePath) == 0 {
		logging.Errorw("Kiva.InitApolloSerializer.PropertiesPath.Error", "error", EnvApolloConfErr)
		logging.Sync()
		os.Exit(1)
	}
	conf := GetConfig(absolutePath)
	if conf == nil {
		logging.Sync()
		os.Exit(1)
	}
	err := apolloStart(conf)
	if err != nil {
		logging.Errorw("Kiva.InitApolloSerializer.ApolloStart.Error", "error", err)
	}
}

// 通过配置初始化 自动序列化
func InitApolloSerializerByConf(conf *agollo.Conf) {
	err := apolloStart(conf)
	if err != nil {
		logging.Errorw("Kiva.InitApolloSerializerByConf.ApolloStart.Error", "error", err)
	}
}

// 通过字符串内容获取配置
func GetConfigByStr(s string) *agollo.Conf {
	logging.Infow("Kiva.GetConfigByBytes.ReadConf.Info", "content", s)
	return GetConfigByBytes([]byte(s))
}

// 通过字符串内容获取配置
func GetConfigByBytes(s []byte) *agollo.Conf {
	logging.Infow("Kiva.GetConfigByBytes.ReadConf.Info", "content", string(s))
	var conf = &agollo.Conf{}
	err := json.Unmarshal(s, conf)

	if err != nil {
		logging.Errorw("Kiva.ReadConf.NewConf.Error", "error", err)
		return nil
	}
	return conf
}

// 通过路径获取配置
func GetConfig(pt string) *agollo.Conf {
	logging.Infow("Kiva.ReadConf.Info", "path", pt)

	conf, err := agollo.NewConf(pt)
	if err != nil {
		logging.Errorw("Kiva.ReadConf.NewConf.Error", "err", err)
		return nil
	}
	return conf
}

// 最基本的运行 apollo 自动序列化
func InitApolloByPath(pt string) {
	conf := GetConfig(pt)
	if conf == nil {
		logging.Sync()
		os.Exit(1)
	}
	err := apolloStart(conf)
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

// 支持动态扩展 namespace, 自动序列化
func NewApolloSerializer(conf *agollo.Conf, handler option) *ApolloSerializer {
	ap := &ApolloSerializer{conf: conf, Client: agollo.NewClient(conf), handler: handler}
	ap.ctx, ap.cancel = context.WithCancel(context.Background())
	return ap
}

// 支持动态扩展 namespace, 自动序列化
func NewApolloSerializerByPath(p string, handler option) *ApolloSerializer {
	conf := GetConfig(p)
	if conf == nil {
		logging.Sync()
		os.Exit(1)
	}
	return NewApolloSerializer(conf, handler)
}

type ApolloSerializer struct {
	conf    *agollo.Conf
	Client  *agollo.Client
	handler option
	ctx     context.Context
	cancel  context.CancelFunc
}

func (ap *ApolloSerializer) SubscribeNameSpaces(ns ...string) {
	_ = ap.Client.SubscribeToNamespaces(ns...)
}

func (ap *ApolloSerializer) upload(ns string) error {

	// 监听更新
	content := ap.Client.GetNameSpaceContent(ns, "")
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
		if ns == "application" {
			return nil
		}
		return errors.Wrapf(SetApolloConfErr, "Namespace '%s'", ns)
	}

	return nil

}
func (ap *ApolloSerializer) ReloadHandler() {
	for _, ns := range ap.conf.NameSpaceNames {
		err := ap.upload(ns)

		if err != nil {
			logging.Warnw("ApolloSerializer.ReloadHandler.Warning", "error", err)
		}

	}

}

// StartWithConf run agollo with Conf
func (ap *ApolloSerializer) Start() {
	err := ap.Client.Start()
	if err != nil {
		logging.Errorw("Kiva.ApolloSerializer.Start.Error", "error", err)
		logging.Sync()
		os.Exit(-1)
	}

	handler := ap.handler

	ap.ReloadHandler()
	go func() {
		for {
			select {
			case ev := <-ap.Client.WatchUpdate():
				logging.Infow("Kiva.ApolloSerializer.Start.WatchUpdate.Info", "namespace", ev.Namespace, )

				if handler != nil {
					err = handler(ev)
				}
				if err != nil {
					logging.Infow("Kiva.ApolloSerializer.Start.WatchHandler.Error", "namespace", ev.Namespace, "error", err)
				}
					err = ap.upload(ev.Namespace)
				if err != nil {
					logging.Infow("Kiva.ApolloSerializer.Start.WatchUpdate.Error", "namespace", ev.Namespace, "error", err)
				}
			case <-ap.ctx.Done():
				logging.Infof("Kiva.ApolloSerializer.Safe.Stop")
				return
			}
		}
	}()
}

// Stop sync config
func (ap *ApolloSerializer) Stop() {
	ap.cancel()
	_ = ap.Client.Stop()
}

// 最基本的运行 apollo 自动加载并维护 watch
func RunApolloByPath(path string, handler watchOption) {
	conf := GetConfig(path)
	if conf == nil {
		logging.Sync()
		os.Exit(1)
	}
	RunApollo(conf, handler)
}

// 最基本的运行 apollo 自动加载并维护 watch
func RunApollo(conf *agollo.Conf, handler watchOption) {
	err := agollo.StartWithConf(conf)
	if err != nil {
		logging.Errorf("start agollo config error:%v", err)
		logging.Sync()
		os.Exit(-1)
	}
	if handler == nil {
		return
	}
	for _, ns := range conf.NameSpaceNames {
		handler(ns, agollo.GetNameSpaceContent(ns, ""))
	}
	go func() {
		for ev := range agollo.WatchUpdate() {
			logging.Infof("apollo changed, namespace:%s", ev.Namespace)
			handler(ev.Namespace, agollo.GetNameSpaceContent(ev.Namespace, ""))
		}
	}()
}
