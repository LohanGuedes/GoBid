package user

import (
	"context"

	"github.com/rocketseat/go-first-auth/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.Email), "user_name", "this field cannot be blank")
	eval.CheckField(validator.NotBlank(req.Password), "password", "this field cannot be blank")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "not a valid email")

	return eval
}
