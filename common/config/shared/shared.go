package shared

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/viper"
	"net"
	"net/netip"
	"reflect"
	"regexp"
)

type MapsetWrapper[T comparable] struct {
	mapset.Set[T]
}

type Interface struct {
	// Regex indicates if Value is a regex pattern.
	Regex bool
	// Value is either a regex pattern or a fixed string.
	Value string `validate:"required,maybe_regex"`
}

func (i *Interface) Matches(value string) bool {
	if i.Regex {
		if regex, err := regexp.Compile(i.Value); err != nil {
			return false
		} else {
			return regex.MatchString(value)
		}
	}
	return i.Value == value
}

func maybeRegex(fl validator.FieldLevel) bool {
	regex := fl.Parent().FieldByName("Regex").Bool()
	if !regex {
		return true
	}
	_, err := regexp.Compile(fl.Field().String())
	return err == nil
}

func ipMulticast(fl validator.FieldLevel) bool {
	return fl.Field().Interface().(netip.Addr).IsMulticast()
}

func ipAddrV4(fl validator.FieldLevel) bool {
	return fl.Field().Interface().(netip.Addr).Is4()
}

func ipAddrV6(fl validator.FieldLevel) bool {
	return fl.Field().Interface().(netip.Addr).Is6()
}

func ipAddrNoZone(fl validator.FieldLevel) bool {
	if ipAddr, err := netip.ParseAddr(fl.Field().String()); err == nil {
		return ipAddr.Zone() == ""
	}
	return true
}

func supportedGameTitle(fl validator.FieldLevel) bool {
	return common.SupportedGameTitles.ContainsOne(fl.Field().Interface().(common.GameTitle))
}

func ConvertSlice[T any](slice []any) ([]T, error) {
	result := make([]T, len(slice))
	for i, v := range slice {
		if val, ok := v.(T); ok {
			result[i] = val
		} else {
			var zeroT T
			return nil, fmt.Errorf("expected a slice of %T, got %T", zeroT, v)
		}
	}
	return result, nil
}

func MapSetWrapperHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.Slice {
			return data, nil
		}
		slice, _ := data.([]any)
		switch t {
		case reflect.TypeOf(MapsetWrapper[string]{}):
			if sliceStrings, err := ConvertSlice[string](slice); err != nil {
				return nil, err
			} else {
				return MapsetWrapper[string]{
					Set: mapset.NewThreadUnsafeSet[string](sliceStrings...),
				}, nil
			}
		case reflect.TypeOf(MapsetWrapper[uint16]{}):
			if sliceInt64, err := ConvertSlice[int64](slice); err != nil {
				return nil, err
			} else {
				sliceUint16 := make([]uint16, len(sliceInt64))
				for i, v := range sliceInt64 {
					if v < 0 || v > 65535 {
						return data, fmt.Errorf("expected a slice of uint16, got %T", v)
					} else {
						sliceUint16[i] = uint16(v)
					}
				}
				return MapsetWrapper[uint16]{
					Set: mapset.NewThreadUnsafeSet[uint16](sliceUint16...),
				}, nil
			}
		case reflect.TypeOf(MapsetWrapper[netip.Addr]{}):
			if sliceString, err := ConvertSlice[string](slice); err != nil {
				return nil, err
			} else {
				sliceNetIpAddr := make([]netip.Addr, len(sliceString))
				for i, v := range sliceString {
					if addr, err := netip.ParseAddr(v); err != nil {
						return data, fmt.Errorf("expected a slice of netip.Addr, got %T", v)
					} else {
						sliceNetIpAddr[i] = addr
					}
				}
				return MapsetWrapper[netip.Addr]{
					Set: mapset.NewThreadUnsafeSet[netip.Addr](sliceNetIpAddr...),
				}, nil
			}
		case reflect.TypeOf(MapsetWrapper[Interface]{}):
			iffs := make([]Interface, len(slice))
			for i, iffRaw := range slice {
				iff := Interface{}
				iffMap := iffRaw.(map[string]any)
				if iffMap == nil {
					return data, fmt.Errorf("expected a slice of Interface, got %T", iffRaw)
				}
				if regex, ok := iffMap["regex"].(bool); ok {
					iff.Regex = regex
				}
				if value, ok := iffMap["value"].(string); ok {
					iff.Value = value
				} else {
					return data, fmt.Errorf("expected a slice of Interface with 'Value', got %T", iffRaw)
				}
				iffs[i] = iff
			}
			return MapsetWrapper[Interface]{
				Set: mapset.NewThreadUnsafeSet[Interface](iffs...),
			}, nil
		case reflect.TypeOf(MapsetWrapper[common.GameTitle]{}):
			if sliceStrings, err := ConvertSlice[string](slice); err != nil {
				return nil, err
			} else {
				sliceGameTitle := make([]common.GameTitle, len(sliceStrings))
				for i, v := range sliceStrings {
					sliceGameTitle[i] = common.GameTitle(v)
				}
				return MapsetWrapper[common.GameTitle]{
					Set: mapset.NewThreadUnsafeSet[common.GameTitle](sliceGameTitle...),
				}, nil
			}
		}

		return data, nil
	}
}

func GenericSetTypeFunc(field reflect.Value) interface{} {
	embeddedSetField := field.Field(0)
	if embeddedSetField.IsValid() && !embeddedSetField.IsNil() {
		concreteSet := embeddedSetField.Elem()
		toSliceMethod := concreteSet.MethodByName("ToSlice")
		results := toSliceMethod.Call([]reflect.Value{})
		if len(results) > 0 {
			return results[0].Interface()
		}
	}
	return []interface{}{}
}

func Unmarshal(v *viper.Viper, config any) error {
	return v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToNetIPAddrHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
				MapSetWrapperHookFunc(),
			),
		),
	)
}

func Validator() (error, *validator.Validate) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterCustomTypeFunc(
		GenericSetTypeFunc,
		MapsetWrapper[string]{},
		MapsetWrapper[uint16]{},
		MapsetWrapper[netip.Addr]{},
		MapsetWrapper[Interface]{},
		MapsetWrapper[common.GameTitle]{},
	)
	if err := validate.RegisterValidation("ip_addr_v4", ipAddrV4); err != nil {
		return err, nil
	}
	if err := validate.RegisterValidation("ip_addr_v6", ipAddrV6); err != nil {
		return err, nil
	}
	if err := validate.RegisterValidation("ip_addr_no_zone", ipAddrNoZone); err != nil {
		return err, nil
	}
	if err := validate.RegisterValidation("ip_multicast", ipMulticast); err != nil {
		return err, nil
	}
	if err := validate.RegisterValidation("maybe_regex", maybeRegex); err != nil {
		return err, nil
	}
	if err := validate.RegisterValidation("supported_game_title", supportedGameTitle); err != nil {
		return err, nil
	}
	return nil, validate
}

func FilterNetworks(networks map[*net.Interface][]*net.IPNet, filter mapset.Set[Interface], IPv4 bool, IPv6 bool, includeIPv6LinkLocal bool) (ipAddrs mapset.Set[netip.Addr]) {
	ipAddrs = mapset.NewThreadUnsafeSet[netip.Addr]()
	actualFilter := filter.Clone()
	if actualFilter.IsEmpty() {
		// If there is no filter, we use all available interfaces
		actualFilter.Add(
			Interface{
				Regex: true,
				Value: ".*",
			},
		)
	}
	if networks == nil {
		var err error
		networks, err = common.RunningNetworkInterfaces(IPv4, IPv6, includeIPv6LinkLocal)
		if err != nil {
			return
		}
	}
	for iff, nets := range networks {
		for filterIff := range actualFilter.Iter() {
			if filterIff.Matches(iff.Name) {
				for _, n := range nets {
					ipAddrs.Add(common.NetIPToNetIPAddr(n.IP))
				}
			}
		}
	}
	return
}
