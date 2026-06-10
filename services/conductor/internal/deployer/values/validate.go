package values

import (
	"fmt"
	"regexp"
)

var (
	dnsLabel   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	varName    = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	hostnameRe = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)
	maxNameLen = 63
	maxHostLen = 253
)

func Validate(env *Env) error {
	for name := range env.Services {
		if !isValidName(name) {
			return fmt.Errorf("invalid service name %q", name)
		}
	}

	for name := range env.Databases.Postgres {
		if !isValidName(name) {
			return fmt.Errorf("invalid database name %q", name)
		}
	}

	for name := range env.Databases.Redis {
		if !isValidName(name) {
			return fmt.Errorf("invalid redis name %q", name)
		}
	}

	for k := range env.SharedVariables {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid shared variable name %q", k)
		}
	}

	for svcName, svc := range env.Services {
		for k := range svc.Env {
			if !isValidVarName(k) {
				return fmt.Errorf("service %q: invalid variable name %q", svcName, k)
			}
		}

		for _, domain := range svc.Domains {
			if !isValidHostname(domain.Host) {
				return fmt.Errorf("service %q: invalid hostname %q", svcName, domain.Host)
			}
		}

		for _, ref := range svc.SharedRefs {
			if _, ok := env.SharedVariables[ref]; !ok {
				return fmt.Errorf("service %q: sharedRef %q has no matching sharedVariable", svcName, ref)
			}
		}

		for envKey, ref := range svc.DatabaseRefs {
			if !isValidVarName(envKey) {
				return fmt.Errorf("service %q: invalid databaseRef env key %q", svcName, envKey)
			}

			if _, ok := env.Databases.Postgres[ref.Database]; !ok {
				return fmt.Errorf("service %q: databaseRef %q points at unknown database %q", svcName, envKey, ref.Database)
			}
		}

	}

	return nil
}

func isValidName(name string) bool {
	return len(name) > 0 && len(name) <= maxNameLen && dnsLabel.MatchString(name)
}

func isValidPort(port int) bool {
	// port == 0 means no port specified.
	return port >= 0 && port <= 65535
}

func isValidVarName(name string) bool {
	return varName.MatchString(name)
}

func isValidHostname(host string) bool {
	return len(host) > 0 && len(host) <= maxHostLen && hostnameRe.MatchString(host)
}
