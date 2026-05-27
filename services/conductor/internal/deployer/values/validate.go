package values

import (
	"fmt"
	"regexp"
)

var (
	dnsLabel    = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	varName     = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	hostnameRe  = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)
	maxNameLen  = 63
	maxHostLen  = 253
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

		for _, host := range svc.Domains {
			if !isValidHostname(host) {
				return fmt.Errorf("service %q: invalid hostname %q", svcName, host)
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

		for envKey, ref := range svc.ServiceRefs {
			if !isValidVarName(envKey) {
				return fmt.Errorf("service %q: invalid serviceRef env key %q", svcName, envKey)
			}

			if _, ok := env.Services[ref.Service]; !ok {
				return fmt.Errorf("service %q: serviceRef %q points at unknown service %q", svcName, envKey, ref.Service)
			}
		}
	}

	return nil
}

func isValidName(name string) bool {
	return len(name) > 0 && len(name) <= maxNameLen && dnsLabel.MatchString(name)
}

func isValidVarName(name string) bool {
	return varName.MatchString(name)
}

func isValidHostname(host string) bool {
	return len(host) > 0 && len(host) <= maxHostLen && hostnameRe.MatchString(host)
}
