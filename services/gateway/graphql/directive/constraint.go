package directive

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-playground/validator/v10"
)

type Constraint struct {
	validator *validator.Validate
}

func New() *Constraint {
	return &Constraint{
		validator: validator.New(),
	}
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
