package http

import (
	"encoding/json"
	"net/http"
)

type authRegisterRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	DisplayName     string `json:"displayName"`
}

type authLoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type authChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type authRevokeSessionRequest struct {
	SessionID string `json:"sessionId"`
}

type authUpdateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatarUrl"`
}

type authUpdatePreferencesRequest struct {
	PreferredLocale   string `json:"preferredLocale"`
	PreferredLanguage string `json:"preferredLanguage"`
}

func decodeAuthRegisterRequest(r *http.Request) (authRegisterRequest, error) {
	var request authRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authRegisterRequest{}, err
	}
	return request, nil
}

func decodeAuthLoginRequest(r *http.Request) (authLoginRequest, error) {
	var request authLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authLoginRequest{}, err
	}
	return request, nil
}

func decodeAuthChangePasswordRequest(r *http.Request) (authChangePasswordRequest, error) {
	var request authChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authChangePasswordRequest{}, err
	}
	return request, nil
}

func decodeAuthRevokeSessionRequest(r *http.Request) (authRevokeSessionRequest, error) {
	var request authRevokeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authRevokeSessionRequest{}, err
	}
	return request, nil
}

func decodeAuthUpdateProfileRequest(r *http.Request) (authUpdateProfileRequest, error) {
	var request authUpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authUpdateProfileRequest{}, err
	}
	return request, nil
}

func decodeAuthUpdatePreferencesRequest(r *http.Request) (authUpdatePreferencesRequest, error) {
	var request authUpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return authUpdatePreferencesRequest{}, err
	}
	return request, nil
}
