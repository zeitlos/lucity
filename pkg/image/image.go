package image

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

type Ref struct {
	Repository string
	Tag        string
	Digest     string
}

func Parse(s string) (Ref, error) {
	if _, err := name.ParseReference(s, name.WeakValidation); err != nil {
		return Ref{}, err
	}

	return decompose(s), nil
}

func decompose(s string) Ref {
	var ref Ref

	if at := strings.LastIndex(s, "@"); at != -1 {
		ref.Digest = s[at+1:]
		s = s[:at]
	}

	slash := strings.LastIndex(s, "/")

	if colon := strings.LastIndex(s, ":"); colon > slash {
		ref.Repository = s[:colon]
		ref.Tag = s[colon+1:]
	} else {
		ref.Repository = s
	}

	return ref
}

func (r Ref) String() string {
	out := r.Repository

	if r.Tag != "" {
		out += ":" + r.Tag
	}

	if r.Digest != "" {
		out += "@" + r.Digest
	}

	return out
}

func (r Ref) Built() bool {
	return r.Tag != "" || r.Digest != ""
}
