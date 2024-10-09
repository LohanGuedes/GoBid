package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lohanguedes/gobid/internal/store/pgstore"
	"github.com/lohanguedes/gobid/internal/usecase/user"
)

func (api *Api) handleSignUpUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := decodeValidJson[user.CreateUserReq](r)
	if err != nil {
		_ = encodeJson(w, r, http.StatusBadRequest, problems)
		return
	}

	id, err := api.UserService.CreateUser(
		r.Context(),
		data.UserName,
		data.Email,
		data.Password,
		data.Bio,
	)
	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, pgstore.ErrDuplicateEmail) {
			_ = encodeJson(w, r, http.StatusBadRequest, map[string]any{
				"error": "invalid data: duplicate email",
			})
			return
		}
	}

	_ = encodeJson(w, r, http.StatusCreated, map[string]any{
		"user_id": id,
	})
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := decodeValidJson[user.LoginUserReq](r)
	if err != nil {
		_ = encodeJson(w, r, http.StatusBadRequest, problems)
		return
	}

	id, err := api.UserService.Authenticate(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, pgstore.ErrInvalidCredentials) {
			_ = encodeJson(w, r, http.StatusUnprocessableEntity, map[string]string{
				"error": "email or password is incorrect",
			})
			return
		}
		_ = encodeJson(w, r, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err = api.Session.RenewToken(r.Context())
	if err != nil {
		_ = encodeJson(w, r, http.StatusInternalServerError, map[string]string{
			"error": "unexpected internal server error",
		})
		return
	}

	api.Session.Put(r.Context(), "authenticatedUserId", id)

	_ = encodeJson(w, r, http.StatusOK, map[string]string{
		"message": "logged in successfully",
	})
}

func (api *Api) handleLogOut(w http.ResponseWriter, r *http.Request) {
	err := api.Session.RenewToken(r.Context())
	if err != nil {
		_ = encodeJson(w, r, http.StatusInternalServerError, map[string]string{
			"error": "unexpected internal server error",
		})
		return
	}

	api.Session.Remove(r.Context(), "authenticatedUserId")
	_ = encodeJson(w, r, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}
