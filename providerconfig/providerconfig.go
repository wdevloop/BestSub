package providerconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"

	"github.com/antonmedv/expr"
	"github.com/bestruirui/bestsub/proxy/info"
	"gopkg.in/yaml.v3"
)

type providerFile struct {
	Providers map[string]Provider `yaml:"providers"`
}

type Rule struct {
	Restriction string `yaml:"restriction"`
	Order       string `yaml:"order"`
	Desc        bool   `yaml:"desc"`
}

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Provider struct {
	Output     string    `yaml:"output"`
	LowerBound int       `yaml:"lowerbound"`
	RuleSets   []RuleSet `yaml:"ruleSets"`
	Rules      []Rule    `yaml:"rules"`
}

// exprEnv exposes proxy info fields and helper functions to expression
type exprEnv struct {
	info.ProxyInfo
	Re   func(string, string) bool
	Size func(any) int
	Sum  func(...any) float64
	Map  func(...any) float64
}

func newEnv(p info.ProxyInfo) exprEnv {
	return exprEnv{
		ProxyInfo: p,
		Re:        regexpMatch,
		Size:      sizeOf,
		Sum:       sum,
		Map:       mapValue,
	}
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
		if len(p.RuleSets) == 0 && len(p.Rules) > 0 {
			p.RuleSets = []RuleSet{{Rules: p.Rules}}
		}
		providers = append(providers, p)
	}
	return providers, nil
}

// Apply executes provider rules on proxies and returns selected proxies
func Apply(p Provider, proxies []info.Proxy) []info.Proxy {
	current := proxies
	var prevOrder string
	var prevDesc bool
	for _, rs := range p.RuleSets {
		for i, r := range rs.Rules {
			filtered := make([]info.Proxy, 0)
			for _, pr := range current {
				if r.Restriction == "" {
					filtered = append(filtered, pr)
					continue
				}
				ok, err := evalBool(r.Restriction, pr.Info)
				if err != nil {
					continue
				}
				if ok {
					filtered = append(filtered, pr)
				}
			}
			sort.Slice(filtered, func(i, j int) bool {
				af, _ := evalFloat(r.Order, filtered[i].Info)
				bf, _ := evalFloat(r.Order, filtered[j].Info)
				if r.Desc {
					return af > bf
				}
				return af < bf
			})
			if p.LowerBound > 0 && len(filtered) < p.LowerBound {
				extras := append([]info.Proxy{}, current...)
				if prevOrder != "" {
					sort.Slice(extras, func(i, j int) bool {
						af, _ := evalFloat(prevOrder, extras[i].Info)
						bf, _ := evalFloat(prevOrder, extras[j].Info)
						if prevDesc {
							return af > bf
						}
						return af < bf
					})
				}
				ids := make(map[int]struct{})
				for _, pr := range filtered {
					ids[pr.Id] = struct{}{}
				}
				for _, pr := range extras {
					if len(filtered) >= p.LowerBound {
						break
					}
					if _, ok := ids[pr.Id]; !ok {
						filtered = append(filtered, pr)
						ids[pr.Id] = struct{}{}
					}
				}
			}
			current = filtered
			prevOrder = r.Order
			prevDesc = r.Desc
			if i == 0 && prevOrder == "" {
				prevOrder = r.Order
				prevDesc = r.Desc
			}
		}
	}
	return current
}

// evaluate boolean expression using expr language
func evalBool(expression string, inf info.ProxyInfo) (bool, error) {
	out, err := expr.Eval(expression, newEnv(inf))
	if err != nil {
		return false, err
	}
	b, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("expression '%s' does not return bool", expression)
	}
	return b, nil
}

// evaluate numeric expression using expr language
func evalFloat(expression string, inf info.ProxyInfo) (float64, error) {
	out, err := expr.Eval(expression, newEnv(inf))
	if err != nil {
		return 0, err
	}
	if f, ok := toFloat(out); ok {
		return f, nil
	}
	return 0, fmt.Errorf("expression '%s' does not return number", expression)
}

func regexpMatch(pattern, value string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

func sizeOf(v any) int {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len()
	default:
		return 0
	}
}

func sum(args ...any) float64 {
	var total float64
	for _, a := range args {
		if f, ok := toFloat(a); ok {
			total += f
		}
	}
	return total
}

func mapValue(args ...any) float64 {
	for i := 0; i+1 < len(args); i += 2 {
		cond, _ := args[i].(bool)
		if cond {
			if f, ok := toFloat(args[i+1]); ok {
				return f
			}
			return 0
		}
	}
	if len(args)%2 == 1 {
		if f, ok := toFloat(args[len(args)-1]); ok {
			return f
		}
	}
	return 0
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

func MarshalExample() ([]byte, error) {
	example := providerFile{
		Providers: map[string]Provider{
			"USFastClean": {
				Output:     "us_fast.yaml",
				LowerBound: 5,
				Rules: []Rule{
					{Restriction: `re("^US$", Country) && Alive`, Order: "Speed", Desc: true},
					{Restriction: "Delay < 50", Order: "Delay", Desc: false},
				},
			},
		},
	}
	return json.MarshalIndent(example, "", "  ")
}
