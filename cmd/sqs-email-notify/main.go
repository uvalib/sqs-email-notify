package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

type MessageTuple struct {
	id            string // message ID
	FirstSent     uint64 // first sent
	FirstReceived uint64 // first received
}

func (mt MessageTuple) ToString() string {
	ts := time.Unix(int64(mt.FirstSent/1000), 0) // cos our format is epoch plus milliseconds
	return fmt.Sprintf("Identifier: %s (first sent: %s)", mt.id, ts)
}

//
// main entry point
//
func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	// get the queue handles from the queue name
	inQueueHandle, err := aws.QueueHandle(cfg.InQueueName)
	fatalIfError(err)

	// our list of unprocessed messages
	var messageList []MessageTuple
	for {
		// are there any messages to be processed
		count, e := getQueueMessageCount(aws, cfg.InQueueName)
		fatalIfError(e)
		if count == 0 {
			log.Printf("INFO: queue %s contains no messages, sleeping for %d minutes", cfg.InQueueName, cfg.WaitTime)
			time.Sleep(time.Duration(cfg.WaitTime) * time.Minute)
			continue
		}

		// we know we have at least count messages
		for {

			// wait for a batch of messages
			messages, err := aws.BatchMessageGet(inQueueHandle, awssqs.MAX_SQS_BLOCK_COUNT, time.Duration(cfg.PollTimeOut)*time.Second)
			fatalIfError(err)

			// did we receive any?
			sz := len(messages)
			if sz != 0 {

				// extract the ID from each message
				for ix := range messages {
					id, found := messages[ix].GetAttribute(awssqs.AttributeKeyRecordId)
					if found == false {
						id = "unknown" // should not typically happen
					}
					messageList = append(messageList, MessageTuple{id, messages[ix].FirstSent, messages[ix].FirstReceived})
				}

				// should we delete these messages (maybe not for testing)
				if cfg.PurgeMessages == true {
					// delete the messages, ignore normal failures as we will get them next time
					_, err := aws.BatchMessageDelete(inQueueHandle, messages)
					if err != nil && err != awssqs.ErrOneOrMoreOperationsUnsuccessful {
						fatalIfError(err)
					}
				}

			} else {
				log.Printf("INFO: no more messages available")
				break
			}
		}

		// we now have a list of ID's to process...
		pending := len(messageList)
		if pending != 0 {

			// if we need to create and attach a file to the email
			if pending > cfg.EmailIdLimit {
				createZipfile(fmt.Sprintf("%s/%s", cfg.TmpDir, cfg.EmailAttachName), messageList)
			}

			// create and send the email
			sendNotificationEmail(cfg, messageList)

			log.Printf("INFO: processing complete (%d messages), sleeping for %d minutes", pending, cfg.WaitTime)
			messageList = messageList[:0]
			time.Sleep(time.Duration(cfg.WaitTime) * time.Minute)
		}
	}
}

//
// end of file
//
