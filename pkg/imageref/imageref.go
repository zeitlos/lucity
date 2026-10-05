package imageref

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

// Validate returns an error when ref is not a valid image reference, or when
// it names an image on the platform registry outside the workspace's own
// repositories. registry is the address the caller's consumer holds
// credentials for. It parses ref with the grammar kubelet, containerd and
// BuildKit use, so the host and path it checks are the ones they pull from.
func Validate(ref, workspace, registry string) error {
	named, err := reference.ParseNormalizedNamed(ref)

	if err != nil {
		return fmt.Errorf("invalid image reference %q: %w", ref, err)
	}

	if strings.EqualFold(reference.Domain(named), registry) && !strings.HasPrefix(reference.Path(named), workspace+"/") {
		return fmt.Errorf("image %q is outside this workspace", ref)
	}

	return nil
}
