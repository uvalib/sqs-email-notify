package main

import (
	"log"
	"os"
	"strconv"
	"text/template"
)

// ServiceConfig defines all of the service configuration parameters
type ServiceConfig struct {
	InQueueName       string // SQS queue name to monitor for message
	MessageBucketName string // the bucket to use for large messages
	PollTimeOut       int64  // the SQS queue timeout (in seconds)
	PurgeMessages     bool   // do we purge processed messages or not

	WaitTime int    // the check wait time (in minutes)
	TmpDir   string // the temporary directory (local)

	SMTPHost string // SMTP hostname
	SMTPPort int    // SMTP port number
	SMTPUser string // SMTP username
	SMTPPass string // SMTP password

	EmailSender     string // the email sender
	EmailRecipient  string // the email recipient
	EmailCC         string // the email CC
	EmailSubject    string // the email subject
	EmailTemplate   string // the email template, run through the template engine
	EmailIdLimit    int    // the maximum number of ID's to list in the email (create attachment if exceeded)
	EmailAttachName string // the name of the file of ID's to attach
	SendEmail       bool   // do we send or just log
}

func envWithDefault(env string, defaultValue string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("INFO: environment variable not set: [%s] using default value [%s]", env, defaultValue)
		return defaultValue
	}

	return val
}

func ensureSet(env string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("FATAL ERROR: environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func ensureSetAndNonEmpty(env string) string {
	val := ensureSet(env)

	if val == "" {
		log.Printf("FATAL ERROR: environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func envToInt(env string) int {

	number := ensureSetAndNonEmpty(env)
	n, err := strconv.Atoi(number)
	fatalIfError(err)
	return n
}

func envToBool(env string) bool {

	str := ensureSetAndNonEmpty(env)
	b, err := strconv.ParseBool(str)
	fatalIfError(err)
	return b
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	cfg.InQueueName = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_IN_QUEUE")
	cfg.MessageBucketName = ensureSetAndNonEmpty("SQS_MESSAGE_BUCKET")
	cfg.PollTimeOut = int64(envToInt("SQS_EMAIL_NOTIFY_QUEUE_POLL_TIMEOUT"))
	cfg.PurgeMessages = envToBool("SQS_EMAIL_NOTIFY_PURGE_MESSAGES")

	cfg.WaitTime = envToInt("SQS_EMAIL_NOTIFY_WAIT_TIME")
	cfg.TmpDir = envWithDefault("SQS_EMAIL_NOTIFY_TMP_DIR", "/tmp")

	cfg.SMTPHost = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_SMTP_HOST")
	cfg.SMTPPort = envToInt("SQS_EMAIL_NOTIFY_SMTP_PORT")
	cfg.SMTPUser = ensureSet("SQS_EMAIL_NOTIFY_SMTP_USER")
	cfg.SMTPPass = ensureSet("SQS_EMAIL_NOTIFY_SMTP_PASSWORD")

	cfg.EmailSender = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_EMAIL_SENDER")
	cfg.EmailRecipient = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_EMAIL_RECIPIENT")
	cfg.EmailCC = ensureSet("SQS_EMAIL_NOTIFY_EMAIL_CC")
	cfg.EmailSubject = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_EMAIL_SUBJECT")
	cfg.EmailTemplate = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_EMAIL_TEMPLATE")
	cfg.EmailIdLimit = envToInt("SQS_EMAIL_NOTIFY_EMAIL_ID_LIMIT")
	cfg.EmailAttachName = ensureSetAndNonEmpty("SQS_EMAIL_NOTIFY_EMAIL_ATTACH_NAME")
	cfg.SendEmail = envToBool("SQS_EMAIL_NOTIFY_EMAIL_SEND")

	log.Printf("[CONFIG] InQueueName       = [%s]", cfg.InQueueName)
	log.Printf("[CONFIG] MessageBucketName = [%s]", cfg.MessageBucketName)
	log.Printf("[CONFIG] PollTimeOut       = [%d]", cfg.PollTimeOut)
	log.Printf("[CONFIG] PurgeMessages     = [%t]", cfg.PurgeMessages)

	log.Printf("[CONFIG] WaitTime          = [%d]", cfg.WaitTime)
	log.Printf("[CONFIG] TmpDir            = [%s]", cfg.TmpDir)

	log.Printf("[CONFIG] SMTPHost          = [%s]", cfg.SMTPHost)
	log.Printf("[CONFIG] SMTPPort          = [%d]", cfg.SMTPPort)
	log.Printf("[CONFIG] SMTPUser          = [%s]", cfg.SMTPUser)
	log.Printf("[CONFIG] SMTPPass          = [%s]", cfg.SMTPPass)

	log.Printf("[CONFIG] EmailSender       = [%s]", cfg.EmailSender)
	log.Printf("[CONFIG] EmailRecipient    = [%s]", cfg.EmailRecipient)
	log.Printf("[CONFIG] EmailCC           = [%s]", cfg.EmailCC)
	log.Printf("[CONFIG] EmailSubject      = [%s]", cfg.EmailSubject)
	log.Printf("[CONFIG] EmailTemplate     = [%s]", cfg.EmailTemplate)
	log.Printf("[CONFIG] EmailIdLimit      = [%d]", cfg.EmailIdLimit)
	log.Printf("[CONFIG] EmailAttachName   = [%s]", cfg.EmailAttachName)
	log.Printf("[CONFIG] SendEmail         = [%t]", cfg.SendEmail)

	// validate the template here
	_, err := template.New("email").Parse(cfg.EmailTemplate)
	fatalIfError(err)

	return &cfg
}
