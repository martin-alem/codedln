package service

import (
	"codedln/shared/http_error"
	"codedln/user_module/model"
	"codedln/user_module/repository"
	"codedln/util/constant"
	"codedln/util/types"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/idtoken"
	"net/http"
	"os"
)

type UserService struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, idToken string, signInWith types.OAuthSignIn) (*model.User, error) {

	switch signInWith {
	case constant.GoogleSignIn:
		payload, err := google(ctx, idToken)
		if err != nil {
			return nil, err
		}

		user, err := s.repo.GetUser(ctx, map[string]any{"email": payload.Email})
		//User does not exist in database
		if user == nil && err == nil {
			return s.repo.CreateUser(ctx, *payload)
		}

		//User exist in database
		if user != nil && err == nil {
			return user, nil
		}

		//An error occurred while trying to fetch user
		return nil, err

	default:
		return nil, http_error.New(http.StatusBadRequest, "invalid oauth type")
	}
}

func (s *UserService) GetUser(ctx context.Context, userId string) (*model.User, error) {
	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, http_error.New(http.StatusBadRequest, "invalid user id")
	}
	filter := map[string]any{"_id": objectId}
	return s.repo.GetUser(ctx, filter)
}

func (s *UserService) DeleteUser(ctx context.Context, userId string) error {
	return s.repo.DeleteUser(ctx, userId)
}

func google(ctx context.Context, idToken string) (*model.User, error) {
	// Validate the ID token and obtain the payload
	payload, err := idtoken.Validate(ctx, idToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		// This error might occur if the token is expired or invalid
		return nil, http_error.New(http.StatusBadRequest, "invalid id token")
	}

	// Extract claims from the payload
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, http_error.New(http.StatusBadRequest, "email claim missing in ID token")
	}

	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok {
		return nil, http_error.New(http.StatusBadRequest, "email_verified claim missing in ID token")
	}

	firstName, ok := payload.Claims["given_name"].(string)
	if !ok {
		return nil, http_error.New(http.StatusBadRequest, "given name claim missing in ID token")
	}

	lastName, ok := payload.Claims["family_name"].(string)
	if !ok {
		return nil, http_error.New(http.StatusBadRequest, "family name claim missing in ID token")
	}

	picture, ok := payload.Claims["picture"].(string)
	if !ok {
		return nil, http_error.New(http.StatusBadRequest, "picture claim missing in ID token")
	}

	// Construct a user object. Adjust according to your User model structure
	user := &model.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Picture:   picture,
		Verified:  emailVerified,
	}

	return user, nil
}
