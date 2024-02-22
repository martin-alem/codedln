package service

import (
	"codedln/shared/http_error"
	"codedln/url_module/model"
	"codedln/url_module/repository"
	"codedln/util/constant"
	"codedln/util/types"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type UrlService struct {
	repo repository.UrlRepository
}

func New(repo repository.UrlRepository) *UrlService {
	return &UrlService{
		repo: repo,
	}
}

func (s *UrlService) CreateUrl(ctx context.Context, originalUrl string, alias string, userId string) (*model.Url, error) {

	var shortUrl string

	if alias != "" {
		exist, err := s.AliasExist(ctx, alias)
		if err != nil {
			return nil, err
		}

		if !exist {
			shortUrl = alias
		} else {
			return nil, http_error.New(http.StatusBadRequest, "alias already exist. try another one")
		}

	} else {
		shortUrl = s.GenerateAlias(originalUrl)

		aliasExist, err := s.AliasExist(ctx, shortUrl)
		if err != nil {
			return nil, err
		}
		if aliasExist {
			found := false
			for i := 0; i < constant.AliasRetry; i++ {
				sUrl := s.GenerateAlias(originalUrl)
				exist, existErr := s.AliasExist(ctx, sUrl)
				if existErr != nil {
					return nil, existErr
				}
				if !exist {
					shortUrl = sUrl
					found = true
					break
				}
			}

			if !found {
				return nil, http_error.New(http.StatusInternalServerError, "unable to generate a short url. please contact support")
			}
		}
	}

	var userIdObj primitive.ObjectID
	var err error
	if userId != "" {
		userIdObj, err = primitive.ObjectIDFromHex(userId)
		if err != nil {
			log.Println(err)
			return nil, http_error.New(http.StatusInternalServerError, "unable to parse user id")
		}
	}
	url := model.Url{
		UserId:      userIdObj,
		OriginalUrl: originalUrl,
		Alias:       shortUrl,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	return s.repo.CreateUrl(ctx, url)
}

func (s *UrlService) CheckAliasExistence(ctx context.Context, alias string) error {

	query := bson.D{{"alias", alias}}
	url, err := s.repo.GetUrl(ctx, query)
	if err != nil {
		return err
	}

	if url != nil {
		return http_error.New(http.StatusBadRequest, "alias exist")
	}

	return nil
}

func (s *UrlService) GetUrls(ctx context.Context, query string, sort types.DateSort, limit int64, userId string) (*types.PaginationResult[model.Url], error) {
	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, http_error.New(http.StatusInternalServerError, "unable to parse user id")
	}
	return s.repo.GetUrls(ctx, query, sort, limit, userIdObj)
}

func (s *UrlService) GetUrl(ctx context.Context, urlId string, userId string) (*model.Url, error) {
	urlIdObj, err := primitive.ObjectIDFromHex(urlId)
	if err != nil {
		return nil, http_error.New(http.StatusInternalServerError, "unable to parse urlId")
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, http_error.New(http.StatusInternalServerError, "unable to parse user id")
	}

	query := bson.D{{"$and", bson.A{
		bson.D{{"_id", urlIdObj}},
		bson.D{{"userId", userIdObj}},
	}}}
	return s.repo.GetUrl(ctx, query)
}

func (s *UrlService) DeleteUrl(ctx context.Context, urlId string, userId string) error {
	urlIdObj, err := primitive.ObjectIDFromHex(urlId)
	if err != nil {
		return http_error.New(http.StatusInternalServerError, "unable to parse urlId")
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return http_error.New(http.StatusInternalServerError, "unable to parse user id")
	}
	return s.repo.DeleteUrl(ctx, urlIdObj, userIdObj)
}

func (s *UrlService) DeleteUrls(ctx context.Context, urlIds []string, userId string) error {

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return http_error.New(http.StatusInternalServerError, "unable to parse user id")
	}

	objectIDs := make([]primitive.ObjectID, len(urlIds))
	for i, id := range urlIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			log.Println(err)
			return http_error.New(http.StatusInternalServerError, "invalid id format")
		}
		objectIDs[i] = oid
	}

	return s.repo.DeleteUrls(ctx, objectIDs, userIdObj)

}

func (s *UrlService) Redirect(ctx context.Context, alias string) (string, error) {
	filter := bson.D{{"alias", alias}}
	url, err := s.repo.GetUrl(ctx, filter)
	if err != nil {
		return "", err
	}

	if url == nil {
		return "", http_error.New(http.StatusBadRequest, "no url found for the alias")
	}

	return url.OriginalUrl, nil
}

func (s *UrlService) AliasExist(ctx context.Context, shortUrl string) (bool, error) {
	filter := bson.D{{"alias", shortUrl}}
	url, err := s.repo.GetUrl(ctx, filter)
	if err != nil {
		log.Println(err)
		return true, err
	}

	if url != nil {
		return true, nil
	}

	return false, nil
}

func (s *UrlService) GenerateAlias(originalUrl string) string {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	h := sha256.New()
	h.Write([]byte(originalUrl))
	hashBytes := h.Sum(nil)

	//To ensure that the output of Base64 encoding is exactly 8 characters long,
	//Base64 encoding converts binary data to an ASCII string format by mapping it onto a 64-character set.
	//Each character in Base64 represents 6 bits of the original data.
	//Therefore, to achieve an 8-character Base64 string, a total of 48 bits of data (since 8 characters times 6 bits per character equals 48 bits).
	//This means 6 bytes of data is required because there are 8 bits in a byte, and 48 bits divided by 8 bits per byte equals 6 bytes.
	//So, by selecting 6 bytes of data for Base64 encoding, it ensures that the encoded output will be precisely 8 characters long.
	selectedBytes := make([]byte, 6)
	for i := range selectedBytes {
		selectedBytes[i] = hashBytes[r.Intn(len(hashBytes))]
	}

	// Encode the selected bytes into Base64
	encoded := base64.StdEncoding.EncodeToString(selectedBytes)

	return encoded
}
