package shared

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/viper"
	"net"
	"net/netip"
	"regexp"
	"slices"
)

type Interface struct {
	// Regex indicates if Value is a regex pattern.
	Regex bool
	// Value is either a regex pattern or a fixed string.
	Value string `validate:"required,maybe_regex"`
	regex *regexp.Regexp
}

func (i *Interface) Matches(value string) bool {
	if i.Regex {
		if i.regex == nil {
			var err error
			i.regex, err = regexp.Compile(i.Value)
			if err != nil {
				return false
			}
		}
		return i.regex.MatchString(value)
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

func ipV4(fl validator.FieldLevel) bool {
	return fl.Field().Interface().(net.IP).To4() != nil
}

func ipMulticast(fl validator.FieldLevel) bool {
	return fl.Field().Interface().(net.IP).IsMulticast()
}

func supportedGameTitle(fl validator.FieldLevel) bool {
	return common.SupportedGameTitles.ContainsOne(fl.Field().Interface().(common.GameTitle))
}

func Unmarshal(v *viper.Viper, config any) error {
	return v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(mapstructure.StringToIPHookFunc()),
		),
	)
}

func Validator() (error, *validator.Validate) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("ip_v4", ipV4); err != nil {
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

func FilterNetworks(networks map[*net.Interface][]*net.IPNet, filterIPs []net.IP, filterInterfaces []Interface, listen bool) (IPs []net.IP) {
	ipsSet := mapset.NewThreadUnsafeSet[[4]byte]()
	var actualFilterIPs []net.IP
	actualFilterInterfaces := filterInterfaces
	if match := func(IP net.IP) bool {
		return IP.IsUnspecified()
	}; slices.ContainsFunc(filterIPs, match) {
		// If listening to 0.0.0.0 let the OS handle it
		if listen {
			actualFilterIPs = slices.DeleteFunc(slices.Clone(filterIPs), match)
			ipsSet.Add(netip.IPv4Unspecified().As4())
		} else {
			// otherwise, empty the filter so all interfaces are used
			actualFilterInterfaces = []Interface{}
		}
	}
	if ipsSet.Cardinality() == 0 && len(actualFilterIPs) == 0 && len(actualFilterInterfaces) == 0 {
		// If there is no filter, we use all available interfaces
		actualFilterInterfaces = []Interface{
			{
				Regex: true,
				Value: ".*",
			},
		}
	}
	if networks == nil {
		var err error
		networks, err = common.IPv4RunningNetworkInterfaces()
		if err != nil {
			return
		}
	}
	for iff, nets := range networks {
		var netMatches func(n *net.IPNet) bool
		if len(actualFilterInterfaces) > 0 && slices.ContainsFunc(actualFilterInterfaces, func(filterIff Interface) bool {
			return filterIff.Matches(iff.Name)
		}) {
			netMatches = func(n *net.IPNet) bool {
				return true
			}
		} else if len(actualFilterIPs) > 0 {
			netMatches = func(n *net.IPNet) bool {
				return slices.ContainsFunc(actualFilterIPs, func(IP net.IP) bool {
					return n.IP.Equal(IP)
				})
			}
		}
		if netMatches != nil {
			for _, n := range nets {
				if netMatches(n) {
					ipsSet.Add([4]byte(n.IP.To4()))
				}
			}
		}
	}
	for IP := range ipsSet.Iter() {
		IPs = append(IPs, IP[:])
	}
	return
}
