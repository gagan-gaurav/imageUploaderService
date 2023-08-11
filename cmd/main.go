package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filename := request.QueryStringParameters["filename"]
	filename += ".jpeg" // add the image extension to it.

	height := request.QueryStringParameters["h"]
	width := request.QueryStringParameters["w"]
	h, err1 := strconv.Atoi(height)
	w, err2 := strconv.Atoi(width)
	if err1 != nil || err2 != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid image dimensions",
		}, nil
	}

	// Decode the base64-encoded binary image data
	imageData, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid image data, Cannot load binary Image data",
		}, nil
	}

	// Convert the image data to an image.Image object
	imageReader := bytes.NewReader(imageData)
	img, _, err := image.Decode(imageReader)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid image data, Cannot convert binary into Image object",
		}, nil
	}

	// Resize the image to hxw pixels
	newImg := resize.Resize(uint(h), uint(w), img, resize.Lanczos3)

	// Encode the resized image to JPEG format
	var resizedImageBuffer bytes.Buffer
	jpeg.Encode(&resizedImageBuffer, newImg, nil)

	resizedImageData := resizedImageBuffer.Bytes()

	// Create an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// Upload the image data to an S3 bucket
	s3Client := s3.New(sess)
	bucketName := "boon-image-uploader-service"

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(resizedImageData),
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	responseBody := map[string]string{
		"message": "ImageUploaderService succfully uploaded the image with dimentions " + strconv.Itoa(h) + "x" + strconv.Itoa(w) + ".",
		"url":     "https://boon-image-uploader-service.s3.ap-south-1.amazonaws.com/" + filename,
	}

	// Marshal the response body into JSON format
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// Return the HTTP response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(handler)
}
