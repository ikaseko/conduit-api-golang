package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"rwa/internal/model"
	"rwa/internal/security"

	"github.com/go-playground/validator/v10"
)

type RequestUser struct {
	User struct {
		Username string `json:"username" validate:"required,min=5"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=5"`
	} `json:"user"`
}

func (h *Handlers) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UserRegisterHandler"

	user := RequestUser{}
	err := json.NewDecoder(r.Body).Decode(&user)
	h.log.Info(op + ": attempting to register user")

	if err != nil {
		h.log.Error(op+": failed to decode request body", "error", err)
		HandleError(w, "Invalid request format", http.StatusUnprocessableEntity)
		return
	}

	err = h.V.Struct(user)
	if err != nil {
		h.log.Error(op+": validation failed", "error", err)
		HandleError(w, err.(validator.ValidationErrors).Error(), http.StatusUnprocessableEntity)
		return
	}

	h.log.Info(op+": processing registration", "username", user.User.Username)
	err = h.UserRepository.RegisterUser(user.User.Username, user.User.Email, user.User.Password)
	if err != nil {
		h.log.Error(op+": failed to register user", "error", err)
		HandleError(w, "Failed to register user", http.StatusUnprocessableEntity)
		return
	}

	NewUser, err := h.UserRepository.GetUserForAuth("", user.User.Username)
	if err != nil {
		h.log.Error(op+": failed to retrieve registered user", "error", err)
		HandleError(w, "Failed to retrieve user data", http.StatusUnprocessableEntity)
		return
	}

	token := security.GenerateToken(NewUser.ID)
	err = h.UserRepository.AddToken(token, NewUser.ID)
	if err != nil {
		h.log.Error(op+": failed to add token to database", "error", err)
		HandleError(w, "Failed to create authentication token", http.StatusUnprocessableEntity)
		return
	}

	response := model.UserResponse{
		Id:        NewUser.ID,
		Email:     NewUser.Email,
		Username:  NewUser.Username,
		Bio:       NewUser.Bio,
		Image:     NewUser.Image,
		CreatedAt: NewUser.CreatedAt,
		UpdatedAt: NewUser.UpdatedAt,
		Token:     token,
	}
	responseJSON := model.UserResponseJSON{User: response}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseJSON)
	h.log.Info(op+": user registered successfully", "username", user.User.Username)
	return
}

func (h *Handlers) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetUserHandler"

	ctx := r.Context()
	uid := ctx.Value("uid").(string)
	user, err := h.UserRepository.GetUserForApi(uid)
	if err != nil {
		h.log.Error(op+": failed to get user", "error", err, "uid", uid)
		HandleError(w, "Failed to retrieve user data", http.StatusUnprocessableEntity)
		return
	}

	token, err := h.UserRepository.FindTokenByUID(uid)
	if err != nil {
		h.log.Error(op+": failed to get token", "error", err, "uid", uid)
		HandleError(w, "Failed to retrieve authentication token", http.StatusUnprocessableEntity)
		return
	}

	response := model.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token.Token,
	}
	responseJSON := model.UserResponseJSON{User: response}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseJSON)
	return
}

func (h *Handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.LoginUserHandler"

	loginPayload := struct {
		User struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=5"`
		}
	}{}

	err := json.NewDecoder(r.Body).Decode(&loginPayload)
	if err != nil {
		h.log.Error(op+": failed to decode request body", "error", err)
		HandleError(w, "Invalid request format", http.StatusUnprocessableEntity)
		return
	}

	err = h.V.Struct(loginPayload)
	if err != nil {
		h.log.Error(op+": validation failed", "error", err)
		HandleError(w, err.(validator.ValidationErrors).Error(), http.StatusUnprocessableEntity)
		return
	}

	user, err := h.UserRepository.GetUserForAuth(loginPayload.User.Email, "")
	if err != nil {
		h.log.Error(op+": user not found", "error", err, "email", loginPayload.User.Email)
		HandleError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !security.ComparePasswords(loginPayload.User.Password, user.PasswordHash, user.PasswordSalt) {
		h.log.Error(op+": invalid password", "uid", user.ID)
		HandleError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.UserRepository.FindTokenByUID(user.ID)
	if errors.Is(err, errors.New("no result")) {
		token.Token = security.GenerateToken(user.ID)
		err = h.UserRepository.AddToken(token.Token, user.ID)
		if err != nil {
			h.log.Error(op+": failed to add token", "error", err, "uid", user.ID)
			HandleError(w, "Failed to create authentication token", http.StatusUnprocessableEntity)
			return
		}
	} else if err != nil {
		h.log.Error(op+": failed to find token", "error", err, "uid", user.ID)
		HandleError(w, "Failed to retrieve authentication token", http.StatusUnprocessableEntity)
		return
	}

	response := model.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token.Token,
	}
	responseJSON := model.UserResponseJSON{User: response}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseJSON)
	return
}

func (h *Handlers) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateUserHandler"

	updatePayload := struct {
		User struct {
			Email    string `json:"email" `
			Username string `json:"username"`
			Bio      string `json:"bio"`
			Image    string `json:"image"`
		}
	}{}

	err := json.NewDecoder(r.Body).Decode(&updatePayload)
	if err != nil {
		h.log.Error(op+": failed to decode request body", "error", err)
		HandleError(w, "Invalid request format", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	uid := ctx.Value("uid").(string)
	user, err := h.UserRepository.GetUserForUpdate(uid)
	if err != nil {
		h.log.Error(op+": failed to get user for update", "error", err, "uid", uid)
		HandleError(w, "Failed to retrieve user data", http.StatusUnauthorized)
		return
	}

	// Update user fields if provided
	if updatePayload.User.Email != "" {
		user.Email = updatePayload.User.Email
	}
	if updatePayload.User.Username != "" {
		user.Username = updatePayload.User.Username
	}
	if updatePayload.User.Bio != "" {
		user.Bio = updatePayload.User.Bio
	}
	if updatePayload.User.Image != "" {
		user.Image = updatePayload.User.Image
	}

	err = h.UserRepository.UpdateUser(user)
	if err != nil {
		h.log.Error(op+": failed to update user", "error", err, "uid", uid)
		HandleError(w, "Failed to update user data", http.StatusBadRequest)
		return
	}

	token, err := h.UserRepository.FindTokenByUID(user.ID)
	if errors.Is(err, errors.New("no result")) {
		token.Token = security.GenerateToken(user.ID)
		err = h.UserRepository.AddToken(token.Token, user.ID)
		if err != nil {
			h.log.Error(op+": failed to add token after update", "error", err, "uid", user.ID)
			HandleError(w, "Failed to create authentication token", http.StatusUnprocessableEntity)
			return
		}
	} else if err != nil {
		h.log.Error(op+": failed to find token after update", "error", err, "uid", user.ID)
		HandleError(w, "Failed to retrieve authentication token", http.StatusUnprocessableEntity)
		return
	}

	response := model.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token.Token,
	}
	responseJSON := model.UserResponseJSON{User: response}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseJSON)
	return
}

func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.LogoutHandler"

	uid := r.Context().Value("uid").(string)
	token, err := h.UserRepository.FindTokenByUID(uid)
	if err != nil {
		h.log.Error(op+": failed to find token", "error", err, "uid", uid)
		HandleError(w, "Failed to retrieve authentication token", http.StatusUnauthorized)
		return
	}

	err = h.UserRepository.DeleteToken(token.Token)
	if err != nil {
		h.log.Error(op+": failed to delete token", "error", err, "token", token.Token)
		HandleError(w, "Failed to logout user", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
