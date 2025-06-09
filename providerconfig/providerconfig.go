package providerconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/bestruirui/bestsub/proxy/info"
	"gopkg.in/yaml.v3"
)

type providerFile struct {
	Providers map[string]Provider `yaml:"providers"`
}

type Provider struct {
	Output string                 `yaml:"output"`
	Filter map[string]interface{} `yaml:"filter"`
}

func Load(path string) ([]Provider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pf providerFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	providers := make([]Provider, 0, len(pf.Providers))
	for name, p := range pf.Providers {
		if p.Output == "" {
			p.Output = fmt.Sprintf("%s.yaml", name)
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func Match(filter map[string]interface{}, infoData info.ProxyInfo) bool {
	return matchMap(filter, reflect.ValueOf(infoData))
}

func matchMap(filter map[string]interface{}, val reflect.Value) bool {
	for k, cond := range filter {
		f := findFieldOrMap(val, k)
		if !f.IsValid() {
			return false
		}
		if !matchValue(cond, f) {
			return false
		}
	}
	return true
}

func findFieldOrMap(val reflect.Value, key string) reflect.Value {
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return reflect.Value{}
		}
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < typ.NumField(); i++ {
			if strings.EqualFold(typ.Field(i).Name, key) {
				return val.Field(i)
			}
		}
	case reflect.Map:
		mv := val.MapIndex(reflect.ValueOf(key))
		if mv.IsValid() {
			return mv
		}
		for _, mk := range val.MapKeys() {
			if strings.EqualFold(mk.String(), key) {
				return val.MapIndex(mk)
			}
		}
	}
	return reflect.Value{}
}

func matchValue(cond interface{}, val reflect.Value) bool {
	if val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return false
		}
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Bool:
		b, ok := cond.(bool)
		if !ok {
			return false
		}
		return val.Bool() == b
	case reflect.String:
		s, ok := cond.(string)
		if !ok {
			s = fmt.Sprintf("%v", cond)
		}
		re, err := regexp.Compile(s)
		if err != nil {
			return false
		}
		return re.MatchString(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return matchNumber(cond, float64(val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return matchNumber(cond, float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return matchNumber(cond, val.Float())
	case reflect.Map, reflect.Struct:
		m, ok := cond.(map[string]interface{})
		if !ok {
			return false
		}
		return matchMap(m, val)
	default:
		return false
	}
}

func matchNumber(cond interface{}, val float64) bool {
	switch c := cond.(type) {
	case string:
		s := strings.TrimSpace(c)
		if strings.HasPrefix(s, ">=") {
			t, err := strconv.ParseFloat(strings.TrimSpace(s[2:]), 64)
			if err != nil {
				return false
			}
			return val >= t
		}
		if strings.HasPrefix(s, "<=") {
			t, err := strconv.ParseFloat(strings.TrimSpace(s[2:]), 64)
			if err != nil {
				return false
			}
			return val <= t
		}
		if strings.HasPrefix(s, ">") {
			t, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
			if err != nil {
				return false
			}
			return val > t
		}
		if strings.HasPrefix(s, "<") {
			t, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
			if err != nil {
				return false
			}
			return val < t
		}
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "["), "]"), ",")
			if len(parts) == 2 {
				min, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				max, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err1 != nil || err2 != nil {
					return false
				}
				return val >= min && val <= max
			}
		}
		t, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return val == t
		}
		return false
	case int, int8, int16, int32, int64:
		t := reflect.ValueOf(c).Int()
		return val == float64(t)
	case uint, uint8, uint16, uint32, uint64:
		t := reflect.ValueOf(c).Uint()
		return val == float64(t)
	case float32, float64:
		t := reflect.ValueOf(c).Float()
		return val == t
	default:
		return false
	}
}

func MarshalExample() ([]byte, error) {
	example := providerFile{
		Providers: map[string]Provider{
			"USFastClean": {
				Output: "us_fast.yaml",
				Filter: map[string]interface{}{
					"country": "^US$",
					"alive":   true,
					"speed":   ">=10240",
					"delay":   "<=50",
					"rate":    "[0,1.5]",
					"unlock": map[string]interface{}{
						"netflix": true,
						"disney":  true,
						"youtube": true,
						"chatgpt": true,
					},
					"ip": map[string]interface{}{
						"ipUsage": map[string]interface{}{
							"ipInfo": "[0,1]",
						},
						"ipRisk": map[string]interface{}{
							"ipqs": "<=1",
							"dbip": "<=2",
						},
						"ipRiskFactor": map[string]interface{}{
							"ip2Location": map[string]interface{}{
								"hosting": false,
							},
						},
						"ipBanned": map[string]interface{}{
							"banned": 0,
						},
					},
					"net": map[string]interface{}{
						"latency": map[string]interface{}{
							"International": map[string]interface{}{
								"Tokyo": "<=70",
							},
							"ChinaTelecom": map[string]interface{}{
								"Shanghai": "<=150",
							},
						},
						"route": map[string]interface{}{
							"Beijing-ChinaUnicom-TCP": "AS4837",
						},
					},
				},
			},
			"CNQuality": {
				Filter: map[string]interface{}{
					"country": "^CN$",
					"speed":   ">2048",
					"unlock": map[string]interface{}{
						"tiktok": true,
					},
					"ip": map[string]interface{}{
						"ipBanned": map[string]interface{}{
							"banned": 0,
						},
						"ipRisk": map[string]interface{}{
							"ipqs": "<=2",
						},
					},
				},
			},
		},
	}
	return json.MarshalIndent(example, "", "  ")
}
