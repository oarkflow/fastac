package pathmatch

import "errors"

type Option func(p *Path) error

// SetSeperator specifies the seperator of a path, default: "/"
func SetSeperator(sep string) Option {
	return func(p *Path) error {
		p.Seperator = sep
		return nil
	}
}

// SetWildcard specifies the wildcard string
// default: "*"
func SetWildcard(wildcard string) Option {
	return func(p *Path) error {
		p.Wildcard = wildcard
		return nil
	}
}

// SetPrefix specifies the prefix, which identifies a parameterized segments
// default: ":"
func SetPrefix(prefix string) Option {
	return func(p *Path) error {
		if prefix == "" {
			return errors.New("pathmatch: prefix can't be an empty string")
		}
		p.Prefix = prefix
		return nil
	}
}

// SetPrefix specifies the suffix, which identifies a parameterized segments
// default: ""
func SetSuffix(suffix string) Option {
	return func(p *Path) error {
		p.Suffix = suffix
		return nil
	}
}

// EnableEqualityCheck enables the equality check between parameterized segments with the same name
// e.g. /foo/:id/bar/:id will not match /foo/1/bar/2, if the equality check is enabled
// default: false
func EnableEqualityCheck(b bool) Option {
	return func(p *Path) error {
		p.equalCheck = b
		return nil
	}
}
