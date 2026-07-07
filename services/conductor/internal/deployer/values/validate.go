package values

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	dnsLabel     = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	varName      = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	hostnameRe   = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$`)
	labelValueRe = regexp.MustCompile(`^[a-z0-9A-Z]([a-z0-9A-Z._-]*[a-z0-9A-Z])?$`)
	labelKeyRe   = regexp.MustCompile(`^([a-z0-9A-Z]([a-z0-9A-Z.-]*[a-z0-9A-Z])?/)?[a-z0-9A-Z]([a-z0-9A-Z._-]*[a-z0-9A-Z])?$`)

	imageRepoRe   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/:@-]*$`)
	imageTagRe    = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9._-]*$`)
	imageDigestRe = regexp.MustCompile(`^[a-zA-Z0-9]+:[a-zA-Z0-9]+$`)

	mountPathRe = regexp.MustCompile(`^/[a-zA-Z0-9._/-]+$`)
	gitRefRe    = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

	maxNameLen    = 63
	maxHostLen    = 253
	maxKeyLen     = 253
	maxImageLen   = 512
	maxTagLen     = 128
	maxCommandLen = 4096
	maxBranchLen  = 250
)

func Validate(env *Env) error {
	if err := validateLabels("commonLabels", env.CommonLabels); err != nil {
		return err
	}

	if err := validateAnnotationKeys("commonAnnotations", env.CommonAnnotations); err != nil {
		return err
	}

	for name := range env.Services {
		if !isValidName(name) {
			return fmt.Errorf("invalid service name %q", name)
		}
	}

	for name, pg := range env.Databases.Postgres {
		if !isValidName(name) {
			return fmt.Errorf("invalid database name %q", name)
		}

		if err := validateLabels(fmt.Sprintf("database %q labels", name), pg.Labels); err != nil {
			return err
		}

		if err := validateAnnotationKeys(fmt.Sprintf("database %q annotations", name), pg.Annotations); err != nil {
			return err
		}
	}

	for name, vk := range env.Databases.Valkey {
		if !isValidName(name) {
			return fmt.Errorf("invalid key-value store name %q", name)
		}

		if err := validateLabels(fmt.Sprintf("key-value store %q labels", name), vk.Labels); err != nil {
			return err
		}

		if err := validateAnnotationKeys(fmt.Sprintf("key-value store %q annotations", name), vk.Annotations); err != nil {
			return err
		}
	}

	for name, vol := range env.Volumes {
		if !isValidName(name) {
			return fmt.Errorf("invalid volume name %q", name)
		}

		if err := validateLabels(fmt.Sprintf("volume %q labels", name), vol.Labels); err != nil {
			return err
		}

		if err := validateAnnotationKeys(fmt.Sprintf("volume %q annotations", name), vol.Annotations); err != nil {
			return err
		}

		if _, err := resource.ParseQuantity(vol.Size); err != nil {
			return fmt.Errorf("volume %q: invalid size %q: %w", name, vol.Size, err)
		}
	}

	for k := range env.SharedVariables {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid shared variable name %q", k)
		}
	}

	for svcName, svc := range env.Services {
		if err := validateImage(svcName, svc.Image); err != nil {
			return err
		}

		if err := validateLabels(fmt.Sprintf("service %q labels", svcName), svc.Labels); err != nil {
			return err
		}

		if err := validateLabels(fmt.Sprintf("service %q podLabels", svcName), svc.PodLabels); err != nil {
			return err
		}

		if err := validateAnnotationKeys(fmt.Sprintf("service %q annotations", svcName), svc.Annotations); err != nil {
			return err
		}

		if err := validateAnnotationKeys(fmt.Sprintf("service %q podAnnotations", svcName), svc.PodAnnotations); err != nil {
			return err
		}

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

		for envKey, ref := range svc.Refs {
			if !isValidVarName(envKey) {
				return fmt.Errorf("service %q: invalid ref env key %q", svcName, envKey)
			}

			if ref.Secret == "" || ref.Key == "" {
				return fmt.Errorf("service %q: ref %q must set secret and key", svcName, envKey)
			}
		}

		for volName, path := range svc.VolumeMounts {
			if _, ok := env.Volumes[volName]; !ok {
				return fmt.Errorf("service %q: mounts unknown volume %q", svcName, volName)
			}

			if !isValidMountPath(path) {
				return fmt.Errorf("service %q: invalid mount path %q for volume %q", svcName, path, volName)
			}
		}

	}

	return nil
}

func isValidName(name string) bool {
	return len(name) > 0 && len(name) <= maxNameLen && dnsLabel.MatchString(name)
}

func validateStartCommand(command string) error {
	if len(command) > maxCommandLen {
		return fmt.Errorf("start command exceeds %d characters", maxCommandLen)
	}

	for _, r := range command {
		if r == '\t' {
			continue
		}

		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("start command contains a control character")
		}
	}

	return nil
}

func validateBranch(branch string) error {
	if branch == "" {
		return nil
	}

	if len(branch) > maxBranchLen {
		return fmt.Errorf("branch exceeds %d characters", maxBranchLen)
	}

	if !gitRefRe.MatchString(branch) {
		return fmt.Errorf("invalid branch name %q", branch)
	}

	if strings.HasPrefix(branch, "-") ||
		strings.HasPrefix(branch, "/") ||
		strings.HasSuffix(branch, "/") ||
		strings.HasSuffix(branch, ".") ||
		strings.HasSuffix(branch, ".lock") ||
		strings.Contains(branch, "..") ||
		strings.Contains(branch, "//") {
		return fmt.Errorf("invalid branch name %q", branch)
	}

	return nil
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

func isValidMountPath(path string) bool {
	if len(path) < 2 || len(path) > 256 {
		return false
	}

	if !mountPathRe.MatchString(path) {
		return false
	}

	for _, segment := range strings.Split(path, "/") {
		if segment == ".." {
			return false
		}
	}

	return true
}

func validateImage(svcName string, image ImageRef) error {
	if image.Repository == "" {
		return nil
	}

	if len(image.Repository) > maxImageLen || !imageRepoRe.MatchString(image.Repository) {
		return fmt.Errorf("service %q: invalid image repository %q", svcName, image.Repository)
	}

	if image.Tag != "" && (len(image.Tag) > maxTagLen || !imageTagRe.MatchString(image.Tag)) {
		return fmt.Errorf("service %q: invalid image tag %q", svcName, image.Tag)
	}

	if image.Digest != "" && !imageDigestRe.MatchString(image.Digest) {
		return fmt.Errorf("service %q: invalid image digest %q", svcName, image.Digest)
	}

	return nil
}

func validateLabels(scope string, m map[string]string) error {
	for k, v := range m {
		if !isValidLabelKey(k) {
			return fmt.Errorf("%s: invalid label key %q", scope, k)
		}

		if !isValidLabelValue(v) {
			return fmt.Errorf("%s: invalid label value %q for key %q", scope, v, k)
		}
	}

	return nil
}

func validateAnnotationKeys(scope string, m map[string]string) error {
	for k := range m {
		if !isValidLabelKey(k) {
			return fmt.Errorf("%s: invalid annotation key %q", scope, k)
		}
	}

	return nil
}

func isValidLabelKey(key string) bool {
	return len(key) > 0 && len(key) <= maxKeyLen && labelKeyRe.MatchString(key)
}

func isValidLabelValue(value string) bool {
	return len(value) <= maxNameLen && (value == "" || labelValueRe.MatchString(value))
}
