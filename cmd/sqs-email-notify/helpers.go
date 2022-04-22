package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
)

func fatalIfError(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s", err.Error())
	}
}

func getQueueMessageCount(aws awssqs.AWS_SQS, queue string) (uint, error) {

	count, err := aws.GetMessagesAvailable(queue)
	if err != nil {
		return 0, err
	}
	return count, nil
}

//
// end of file
//
