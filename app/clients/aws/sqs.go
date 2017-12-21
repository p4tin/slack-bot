package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SqsClient struct {
	sqsService  *sqs.SQS
	awsClientID string
	awsServer   string
	awsRegion   string
}

func CreateSqsClient(server, region, clientID string) *SqsClient {
	return &SqsClient{
		sqsService:  sqs.New(session.New(), &aws.Config{Region: aws.String("us-east-1")}),
		awsClientID: clientID,
		awsServer:   server,
		awsRegion:   region,
	}
}

func (client *SqsClient) GetSQSQueueDepth(queueName string) string {
	queueUrl := "https://sqs." + client.awsRegion + ".amazonaws.com/" + client.awsClientID + "/" + queueName
	attrib := "ApproximateNumberOfMessages"
	sendParams := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueUrl), // Required
		AttributeNames: []*string{
			&attrib, // Required
		},
	}
	resp2, sendErr := client.sqsService.GetQueueAttributes(sendParams)
	if sendErr != nil {
		fmt.Println("Depth: " + sendErr.Error())
		return "-1"
	}
	return *resp2.Attributes["ApproximateNumberOfMessages"]
}
