package user

import (
	"context"

	"github.com/rocketseat/go-first-auth/internal/validator"
)

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.Email), "user_name", "this field cannot be blank")
	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "this field cannot be blank")
	eval.CheckField(validator.NotBlank(req.Bio), "bio", "this field cannot be blank")
	eval.CheckField(
		validator.MinChars(req.Bio, 10) &&
			validator.MaxChars(req.Bio, 255),
		"bio",
		"this field must have a length > 30 and be < 255")

	eval.CheckField(validator.MinChars(req.Password, 8), "password", "password should be bigger than 8 chars")

	// If field email is blank will already trigger that its not an valid email
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "not a valid email")

	return eval
}
