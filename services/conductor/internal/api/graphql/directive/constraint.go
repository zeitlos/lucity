package directive

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-playground/validator/v10"

	"github.com/zeitlos/lucity/pkg/usererror"
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
		return nil, usererror.Errorf("invalid value for %s", graphql.GetPathContext(ctx).Path())
	}

	if val == nil {
		return val, nil
	}
	if rv := reflect.ValueOf(val); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return val, nil
	}

	if err = c.validator.Var(val, constraint); err != nil {
		return val, usererror.New(describe(fieldName(ctx), err))
	}

	return val, nil
}

func fieldName(ctx context.Context) string {
	for path := graphql.GetPathContext(ctx); path != nil; path = path.Parent {
		if path.Field != nil {
			return *path.Field
		}
	}

	return "value"
}

func describe(field string, err error) string {
	var errs validator.ValidationErrors

	if !errors.As(err, &errs) || len(errs) == 0 {
		return field + " is not valid"
	}

	fe := errs[0]

	unit := ""

	switch fe.Kind() {
	case reflect.String:
		unit = " characters"
	case reflect.Slice, reflect.Array, reflect.Map:
		unit = " items"
	}

	if fe.Param() == "1" {
		unit = strings.TrimSuffix(unit, "s")
	}

	switch fe.Tag() {
	case "min":
		return fmt.Sprintf("%s must be at least %s%s", field, fe.Param(), unit)
	case "max":
		return fmt.Sprintf("%s must be at most %s%s", field, fe.Param(), unit)
	case "resource_name":
		return field + " may only use lowercase letters, digits and hyphens, and must start and end with a letter or digit"
	default:
		return field + " is not valid"
	}
}
