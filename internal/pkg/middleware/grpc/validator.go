// ABOUTME: gRPC request validation interceptor.
// ABOUTME: Validates protobuf messages using go-playground/validator.

package interceptor

import (
	"context"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var validate = validator.New()

// Validator returns a gRPC unary interceptor that validates requests.
func Validator(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := validate.Struct(req); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			return nil, status.Errorf(codes.InvalidArgument, "validation failed: %s", formatValidationErrors(validationErrs))
		}

		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %s", err.Error())
	}

	return handler(ctx, req)
}

// formatValidationErrors formats validation errors into a readable string.
func formatValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return ""
	}

	// Return first error for simplicity
	e := errs[0]

	return e.Field() + " " + e.Tag()
}
