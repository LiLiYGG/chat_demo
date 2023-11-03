package conf

import (
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// yaml文件是一种配置文件
// Dictinary 字典它是json的一个超集，解析出来的内容在pgo中可以解析为结构体。
var Dictinary *map[interface{}]interface{}

// LoadLocales 读取国际化文件
func LoadLocales(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	m := make(map[interface{}]interface{})
	// yaml.Unmarshal 函数负责将 YAML 格式文本解析
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}
	Dictinary = &m
	return nil
}

// T 翻译
// 把返回的提示翻译成对应的语句
func T(key string) string {
	dic := *Dictinary
	keys := strings.Split(key, ".")
	for index, path := range keys {
		// 如果到达了最后一层，寻找目标翻译
		if len(keys) == (index + 1) {
			for k, v := range dic {
				if k, ok := k.(string); ok {
					if k == path {
						if value, ok := v.(string); ok {
							return value
						}
					}
				}
			}
			return path
		}
		// 如果还有下一层，继续寻找
		for k, v := range dic {
			if ks, ok := k.(string); ok {
				if ks == path {
					if dic, ok = v.(map[interface{}]interface{}); ok == false {
						return path
					}
				}
			} else {
				return ""
			}
		}
	}
	return ""
}
