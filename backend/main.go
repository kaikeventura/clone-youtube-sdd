package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"clone-youtube-backend/api"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type VideoServer struct {
	s3Client    *s3.Client
	dynamoClient *dynamodb.Client
}

func (s *VideoServer) CreateUploadUrl(ctx echo.Context) error {
	var req api.UploadRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Dados de entrada inválidos",
		})
	}

	videoID := uuid.New()
	key := "uploads/" + videoID.String() + getExtension(req.FileType)

	presignClient := s3.NewPresignClient(s.s3Client)
	presignResult, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(string(req.FileType)),
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "S3_ERROR",
			Message: "Erro interno ao gerar a URL",
		})
	}

	expiresAt := time.Now().Add(time.Hour)
	uploadURL := strings.Replace(presignResult.URL, endpoint, "/s3", 1)
	return ctx.JSON(http.StatusOK, api.UploadResponse{
		UploadUrl: uploadURL,
		VideoId:   videoID,
		ExpiresIn: 3600,
		ExpiresAt: &expiresAt,
	})
}

func (s *VideoServer) CreateVideo(ctx echo.Context) error {
	var req api.CreateVideoRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Dados de entrada inválidos",
		})
	}

	now := time.Now().UTC()

	item := map[string]dynamodbtypes.AttributeValue{
		"id":         &dynamodbtypes.AttributeValueMemberS{Value: req.Id.String()},
		"titulo":     &dynamodbtypes.AttributeValueMemberS{Value: req.Titulo},
		"url_s3":     &dynamodbtypes.AttributeValueMemberS{Value: req.UrlS3},
		"autor":      &dynamodbtypes.AttributeValueMemberS{Value: req.Autor},
		"createdAt":  &dynamodbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		"updatedAt":  &dynamodbtypes.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
	}

	if req.Descricao != nil {
		item["descricao"] = &dynamodbtypes.AttributeValueMemberS{Value: *req.Descricao}
	}
	if req.ThumbnailUrl != nil {
		item["thumbnailUrl"] = &dynamodbtypes.AttributeValueMemberS{Value: *req.ThumbnailUrl}
	}

	_, err := s.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "DYNAMO_ERROR",
			Message: "Erro interno ao registrar o vídeo",
		})
	}

	video := api.Video{
		Id:           req.Id,
		Titulo:       req.Titulo,
		Descricao:    req.Descricao,
		UrlS3:        req.UrlS3,
		Autor:        req.Autor,
		ThumbnailUrl: req.ThumbnailUrl,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return ctx.JSON(http.StatusCreated, api.VideoResponse{Video: video})
}

func (s *VideoServer) ListVideos(ctx echo.Context, params api.ListVideosParams) error {
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	if params.Limit != nil {
		scanInput.Limit = aws.Int32(int32(*params.Limit))
	}
	if params.NextToken != nil {
		scanInput.ExclusiveStartKey = map[string]dynamodbtypes.AttributeValue{
			"id": &dynamodbtypes.AttributeValueMemberS{Value: *params.NextToken},
		}
	}

	result, err := s.dynamoClient.Scan(context.TODO(), scanInput)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "DYNAMO_ERROR",
			Message: "Erro interno ao listar vídeos",
		})
	}

	var videos []api.Video
	for _, item := range result.Items {
		var video api.Video
		if v, ok := item["id"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.Id, _ = uuid.Parse(v.Value)
		}
		if v, ok := item["titulo"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.Titulo = v.Value
		}
		if v, ok := item["url_s3"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.UrlS3 = v.Value
		}
		if v, ok := item["autor"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.Autor = v.Value
		}
		if v, ok := item["descricao"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.Descricao = &v.Value
		}
		if v, ok := item["thumbnailUrl"].(*dynamodbtypes.AttributeValueMemberS); ok {
			video.ThumbnailUrl = &v.Value
		}
		if v, ok := item["createdAt"].(*dynamodbtypes.AttributeValueMemberS); ok {
			if t, err := time.Parse(time.RFC3339, v.Value); err == nil {
				video.CreatedAt = t
			}
		}
		if v, ok := item["updatedAt"].(*dynamodbtypes.AttributeValueMemberS); ok {
			if t, err := time.Parse(time.RFC3339, v.Value); err == nil {
				video.UpdatedAt = t
			}
		}
		videos = append(videos, video)
	}

	var nextToken *string
	if result.LastEvaluatedKey != nil {
		if v, ok := result.LastEvaluatedKey["id"].(*dynamodbtypes.AttributeValueMemberS); ok {
			nextToken = &v.Value
		}
	}

	return ctx.JSON(http.StatusOK, api.VideoListResponse{
		Videos:    videos,
		NextToken: nextToken,
	})
}

func (s *VideoServer) GetVideoById(ctx echo.Context, id openapi_types.UUID) error {
	result, err := s.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]dynamodbtypes.AttributeValue{
			"id": &dynamodbtypes.AttributeValueMemberS{Value: id.String()},
		},
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "DYNAMO_ERROR",
			Message: "Erro interno ao buscar o vídeo",
		})
	}

	if result.Item == nil {
		return ctx.JSON(http.StatusNotFound, api.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Vídeo não encontrado",
		})
	}

	video := api.Video{
		Id:    id,
		Titulo: result.Item["titulo"].(*dynamodbtypes.AttributeValueMemberS).Value,
		UrlS3: result.Item["url_s3"].(*dynamodbtypes.AttributeValueMemberS).Value,
		Autor: result.Item["autor"].(*dynamodbtypes.AttributeValueMemberS).Value,
	}

	if v, ok := result.Item["descricao"].(*dynamodbtypes.AttributeValueMemberS); ok {
		video.Descricao = &v.Value
	}
	if v, ok := result.Item["thumbnailUrl"].(*dynamodbtypes.AttributeValueMemberS); ok {
		video.ThumbnailUrl = &v.Value
	}
	if v, ok := result.Item["createdAt"].(*dynamodbtypes.AttributeValueMemberS); ok {
		if t, err := time.Parse(time.RFC3339, v.Value); err == nil {
			video.CreatedAt = t
		}
	}
	if v, ok := result.Item["updatedAt"].(*dynamodbtypes.AttributeValueMemberS); ok {
		if t, err := time.Parse(time.RFC3339, v.Value); err == nil {
			video.UpdatedAt = t
		}
	}

	return ctx.JSON(http.StatusOK, api.VideoResponse{Video: video})
}

func getExtension(fileType api.UploadRequestFileType) string {
	switch fileType {
	case api.Videomp4:
		return ".mp4"
	case api.Videowebm:
		return ".webm"
	case api.Videoquicktime:
		return ".mov"
	case api.VideoxMsvideo:
		return ".avi"
	default:
		return ".mp4"
	}
}

func main() {
	s3Client := InitS3Client()
	dynamoClient := InitDynamoDBClient()
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	})

	api.RegisterHandlers(e, &VideoServer{s3Client: s3Client, dynamoClient: dynamoClient})
	e.Logger.Fatal(e.Start(":3000"))
}
