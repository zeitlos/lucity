package directive

import (
	"context"
	"fmt"
	"regexp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-playground/validator/v10"
)

// resourceNamePattern matches valid resource names: lowercase alphanumeric with hyphens,
// must start and end with [a-z0-9], 1-63 characters total.
var resourceNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

type Constraint struct {
	validator *validator.Validate
}

func New() *Constraint {
	v := validator.New()
	v.RegisterValidation("resource_name", func(fl validator.FieldLevel) bool {
		return resourceNamePattern.MatchString(fl.Field().String())
	})
	return &Constraint{validator: v}
}

func (c *Constraint) Validate(ctx context.Context, obj interface{}, next graphql.Resolver, constraint string) (interface{}, error) {
	val, err := next(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid value for %s", graphql.GetPathContext(ctx).Path())
	}

	path := graphql.GetPathContext(ctx).Path()

	if err = c.validator.Var(val, constraint); err != nil {
		return val, fmt.Errorf("value '%s' for %s does not match %s", val, path, constraint)
	}

	return val, nil
}
