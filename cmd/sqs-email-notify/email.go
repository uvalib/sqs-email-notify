package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"text/template"
	"time"
)

func sendNotificationEmail(cfg *ServiceConfig, messageList []MessageTuple) {

	mail := gomail.NewMessage()
	mail.SetHeader("MIME-version", "1.0")
	mail.SetHeader("Content-Type", "text/plain; charset=\"UTF-8\"")
	mail.SetHeader("Subject", cfg.EmailSubject)
	mail.SetHeader("To", cfg.EmailRecipient)
	mail.SetHeader("From", cfg.EmailSender)

	if cfg.EmailCC[0] != "" {
		mail.SetHeader("Cc", cfg.EmailCC...)
	}
	// render the email body
	body, err := renderEmailBody(cfg, messageList)
	fatalIfError(err)
	mail.SetBody("text/plain", body)

	if cfg.SendEmail == false {
		log.Printf("INFO: Email is in debug mode. Logging message instead of sending")
		log.Printf("INFO: ==========================================================")
		mail.WriteTo(log.Writer())
		log.Printf("INFO: ==========================================================")
		return
	}

	// do we need to attach the file if id's?
	if len(messageList) > cfg.EmailIdLimit {
		mail.Attach(fmt.Sprintf("%s/%s", cfg.TmpDir, cfg.EmailAttachName))
	}

	var dialer gomail.Dialer
	log.Printf("INFO: sending '%s' email to '%s'", cfg.EmailSubject, cfg.EmailRecipient)
	if cfg.SMTPPass != "" {
		log.Printf("INFO: sending email with auth")
		dialer = gomail.Dialer{Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUser, Password: cfg.SMTPPass}
	} else {
		log.Printf("INFO: sending email with no auth")
		dialer = gomail.Dialer{Host: cfg.SMTPHost, Port: cfg.SMTPPort}
	}
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err = dialAndSend(dialer, mail)
	fatalIfError(err)
}

func renderEmailBody(cfg *ServiceConfig, messageList []MessageTuple) (string, error) {

	type EmailAttributes struct {
		Recipient   string
		FailedCount int
		Details     string
	}

	// parse the template
	tmpl, err := template.New("email").Parse(cfg.EmailTemplate)
	if err != nil {
		return "", err
	}

	attribs := EmailAttributes{Recipient: cfg.EmailRecipient, FailedCount: len(messageList)}

	// do we need to include a list of ID's in the message body
	if attribs.FailedCount <= cfg.EmailIdLimit {
		var bodyBuffer bytes.Buffer
		bodyBuffer.WriteString("Problems encountered with the following item(s):\n\n")
		for ix := range messageList {
			bodyBuffer.WriteString(fmt.Sprintf("   %s\n", messageList[ix].ToString()))
		}
		attribs.Details = bodyBuffer.String()
	} else {
		attribs.Details = "Please see the attached file for more details.\n"
	}

	var renderedBuffer bytes.Buffer
	err = tmpl.Execute(&renderedBuffer, attribs)
	if err != nil {
		return "", err
	}

	return renderedBuffer.String(), nil
}

func dialAndSend(dialer gomail.Dialer, mail *gomail.Message) error {

	retryCount := 3
	retrySleepTime := 1 * time.Second
	currentCount := 0
	for {
		err := dialer.DialAndSend(mail)
		if err == nil {
			return nil
		}
		currentCount++

		// break when tried too many times
		if currentCount >= retryCount {
			log.Printf("ERROR: email send failed with error (%s), giving up", err)
			return err
		}

		log.Printf("WARNING: email send failed with error (%s), retrying...", err)

		// sleep for a bit before retrying
		time.Sleep(retrySleepTime)
	}
}

//
// end of file
//
