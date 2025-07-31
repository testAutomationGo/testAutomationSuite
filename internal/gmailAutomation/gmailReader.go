package gmailAutomation

import (
	"io"
	"strings"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"

	"log"
)

func GetRegistrationLink(creds, email string) string {
	emailBody := RetryReadGmailReturnBody(creds, email)
	hrefIndex := strings.Index(emailBody, "href=\"") + 6
	trimEmailBody := emailBody[hrefIndex:]
	hrefIndex2 := strings.Index(trimEmailBody, "href=\"") + 6
	trimEmailBody2 := trimEmailBody[hrefIndex2:]
	endLinkIndex := strings.Index(trimEmailBody2, "\" style")
	regLink := trimEmailBody2[:endLinkIndex]
	return regLink
}

func RetryReadGmailReturnBody(creds, registeringEmail string) string {
	emailBody := ReadGmailReturnBody(creds, registeringEmail)
	for i := 0; i < 10; i++ {
		if emailBody == "" {
			testingToolkit.DelaySeconds(2)
			emailBody = ReadGmailReturnBody(creds, registeringEmail)
		}
	}
	return emailBody
}

func ReadGmailReturnBody(creds, registeringEmail string) string {

	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	if err := c.Login("mark@pinata.cloud", creds); err != nil {
		log.Fatal(err)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	seqset := new(imap.SeqSet)
	if mbox.Messages > 10 {
		seqset.AddRange(mbox.Messages-10, mbox.Messages)
	} else {
		seqset.Add("1:*")
	}

	items := []imap.FetchItem{imap.FetchEnvelope, "BODY[]"}
	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqset, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	emailBody := ""
	for msg := range messages {
		toEmail := msg.Envelope.To[0].Address()
		if strings.Contains(toEmail, strings.ToLower(registeringEmail)) {
			section := &imap.BodySectionName{}
			r := msg.GetBody(section)
			if r == nil {
				log.Fatal("Server didn't return message body")
			}

			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Fatal(err)
			}

			for {
				p, err := mr.NextPart()
				if err != nil {
					break
				}

				switch p.Header.(type) {
				case *mail.InlineHeader:

					body, err := io.ReadAll(p.Body)
					if err != nil {
						log.Fatal(err)
					}
					emailBody = string(body)
				}
			}
		}
	}
	return emailBody
}
