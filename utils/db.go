package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DB struct {
	DynamoDB  *dynamodb.DynamoDB
	TableName string
}

func (DB *DB) UpdateRecordSuccess(recordId string, resultUrl string) (*dynamodb.UpdateItemOutput, error) {
	attributeValues := map[string]*dynamodb.AttributeValue{
		":r": {
			S: aws.String(resultUrl),
		},
		":s": {
			S: aws.String("SUCCESS"),
		},
	}
	updateExpression := "set resultUrl = :r, recordStatus = :s"

	return DB.updateItem(recordId, attributeValues, updateExpression)
}

func (DB *DB) UpdateRecordFailure(recordId string) (*dynamodb.UpdateItemOutput, error) {
	attributeValues := map[string]*dynamodb.AttributeValue{
		":s": {
			S: aws.String("FAILED"),
		},
	}
	updateExpression := "set recordStatus = :s"

	return DB.updateItem(recordId, attributeValues, updateExpression)
}

func (DB *DB) updateItem(recordId string, attributeValues map[string]*dynamodb.AttributeValue, updateExpression string) (*dynamodb.UpdateItemOutput, error) {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: attributeValues,
		TableName:                 aws.String(DB.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(recordId),
			},
		},
		UpdateExpression: aws.String(updateExpression),
	}

	return DB.DynamoDB.UpdateItem(input)
}
