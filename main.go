package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

type ParseHeader struct {
	From    string
	To      string
	Subject string
}

type DataBody struct {
	Name string
}

func main() {

	to := []string{"ade.ardian@simp.co.id", "adeardian1994@gmail.com"}
	from := "ict.notifications@simp.co.id"
	subject := "Test email with PDF attachment"
	attachmentPath := "./example.pdf"
	dataBody := DataBody{
		Name: "Recipient",
	}

	header := parseHeader(from, to, subject)

	body := parseBody("travis_complete_task.html", dataBody)

	encodedAttachment, err := addAttachment(attachmentPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Construct the email message
	var message bytes.Buffer

	// Header
	message.WriteString("From: " + header.From + "\r\n")
	message.WriteString("To: " + header.To + "\r\n")
	message.WriteString("Subject: " + header.Subject + "\r\n")
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/mixed; boundary=boundary1\r\n\r\n")

	// Text part (HTML)
	message.WriteString("--boundary1\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	message.WriteString(body + "\r\n\r\n")

	// Attachment part (PDF)
	message.WriteString("--boundary1\r\n")
	message.WriteString("Content-Type: application/pdf\r\n")
	message.WriteString("Content-Disposition: attachment; filename=\"" + filepath.Base(attachmentPath) + "\"\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	message.WriteString(encodedAttachment + "\r\n")

	// Close boundary
	message.WriteString("--boundary1--\r\n")

	sendMail(message.Bytes(), from, to)

	fmt.Println("Email sent successfully!")
}

func sendMail(msg []byte, from string, to []string) {
	smtpHost := "172.20.3.13:25"
	// Connect to the SMTP server
	client, err := smtp.Dial(smtpHost)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Set the sender and recipient
	if err := client.Mail(from); err != nil {
		panic(err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			panic(err)
		}
	}

	// Send the email body
	wc, err := client.Data()
	if err != nil {
		panic(err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}

func readHtmlTemplate(templateFilename string) (string, error) {
	// Open the HTML file
	file, err := os.Open("./template/" + templateFilename)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}
	defer file.Close()

	// Read the file contents
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	return string(content), nil
}

func addAttachment(attachmentPath string) (string, error) {
	// Read the attachment file
	attachmentData, err := ioutil.ReadFile(attachmentPath)
	if err != nil {
		fmt.Println("Error reading attachment file:", err)
		return "", err
	}

	// Encode attachment data as base64
	encodedAttachment := base64.StdEncoding.EncodeToString(attachmentData)

	return encodedAttachment, nil
}

func parseHeader(from string, to []string, subject string) ParseHeader {
	return ParseHeader{
		From:    from,
		To:      strings.Join(to, ","),
		Subject: subject,
	}
}

func parseBody(htmlTemplate string, input DataBody) string {
	emailTemplate, _ := readHtmlTemplate(htmlTemplate)

	// Parse HTML template
	t, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		panic(err)
	}

	// Render HTML template
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, input); err != nil {
		panic(err)
	}

	return tpl.String()
}
