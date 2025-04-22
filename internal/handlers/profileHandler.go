package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"rwa/internal/model"
	"rwa/internal/repository"

	"github.com/gorilla/mux"
)

func (h *Handlers) FollowHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.FollowHandler"

	vars := mux.Vars(r)
	userName := vars["username"]
	if userName == "" {
		h.log.Error("username parameter is missing", "op", op)
		http.Error(w, "Username parameter is missing", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	user := ctx.Value("uid").(string)

	h.log.Info("following user", "op", op, "user", user, "userToFollow", userName)

	userToFollow, err := h.UserRepository.GetUserForAuth("", userName)
	if err != nil {
		h.log.Error("failed to get user to follow", "op", op, "username", userName, "error", err)
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User to follow not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.UserRepository.FollowUser(user, userToFollow.ID)
	if err != nil {
		h.log.Error("failed to follow user", "op", op, "user", user, "userToFollowID", userToFollow.ID, "error", err)
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}
	response := model.Profile{
		Id:        userToFollow.ID,
		Username:  userToFollow.Username,
		Bio:       userToFollow.Bio,
		Image:     userToFollow.Image,
		Following: true,
		CreatedAt: userToFollow.CreatedAt,
		UpdatedAt: userToFollow.UpdatedAt,
	}
	resp := model.ProfileResponse{
		Profile: response,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)

	h.log.Info("successfully followed user", "op", op, "user", user, "followedUser", userName)
	return
}

func (h *Handlers) UnFollowHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.UnFollowHandler"

	vars := mux.Vars(r)
	userName := vars["username"]
	if userName == "" {
		h.log.Error("username parameter is missing", "op", op)
		http.Error(w, "Username parameter is missing", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	followerUID := ctx.Value("uid").(string)

	h.log.Info("unfollowing user", "op", op, "user", followerUID, "userToUnfollow", userName)

	followedUser, err := h.UserRepository.GetUserForAuth("", userName)
	if err != nil {
		h.log.Error("failed to get user to unfollow", "op", op, "username", userName, "error", err)
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User to unfollow not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.UserRepository.UnFollowUser(followerUID, followedUser.ID)
	if err != nil {
		h.log.Error("failed to unfollow user", "op", op, "user", followerUID, "userToUnfollowID", followedUser.ID, "error", err)
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}
	response := model.Profile{
		Id:        followedUser.ID,
		Username:  followedUser.Username,
		Bio:       followedUser.Bio,
		Image:     followedUser.Image,
		Following: false,
		CreatedAt: followedUser.CreatedAt,
		UpdatedAt: followedUser.UpdatedAt,
	}
	resp := model.ProfileResponse{
		Profile: response,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)

	h.log.Info("successfully unfollowed user", "op", op, "user", followerUID, "unfollowedUser", userName)
	return
}

func (h *Handlers) CheckProfileHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.CheckProfileHandler"

	vars := mux.Vars(r)
	userName := vars["username"]
	if userName == "" {
		h.log.Error("username parameter is missing", "op", op)
		http.Error(w, "Username parameter is missing", http.StatusBadRequest)
		return
	}

	followerUIDValue := r.Context().Value("uid")
	var followerUID string
	var isAuthenticated bool
	if followerUIDValue != nil {
		followerUID, isAuthenticated = followerUIDValue.(string)
		if isAuthenticated {
			h.log.Info("checking profile", "op", op, "user", followerUID, "profileUser", userName)
		} else {
			h.log.Info("checking profile (unauthenticated)", "op", op, "profileUser", userName)
		}
	} else {
		isAuthenticated = false
		h.log.Info("checking profile (unauthenticated)", "op", op, "profileUser", userName)
	}

	followedUser, err := h.UserRepository.GetUserForAuth("", userName)
	if err != nil {
		h.log.Error("failed to get profile user", "op", op, "username", userName, "error", err)
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "Profile user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isFollowing := false
	if isAuthenticated {
		isFollowingCheck, errCheck := h.UserRepository.CheckFollow(followerUID, followedUser.ID)
		if errCheck != nil {
			h.log.Error("failed to check follow status", "op", op, "user", followerUID, "profileUser", userName, "error", errCheck)
		} else {
			isFollowing = isFollowingCheck
		}
	}

	response := model.Profile{
		Id:        followedUser.ID,
		Username:  followedUser.Username,
		Bio:       followedUser.Bio,
		Image:     followedUser.Image,
		Following: isFollowing,
		CreatedAt: followedUser.CreatedAt,
		UpdatedAt: followedUser.UpdatedAt,
	}
	resp := model.ProfileResponse{
		Profile: response,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)

	h.log.Info("successfully checked profile", "op", op, "profileUser", userName, "isFollowing", isFollowing)
	return
}
