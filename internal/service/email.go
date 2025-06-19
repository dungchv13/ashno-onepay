package service

import (
	"ashno-onepay/internal/config"
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/skip2/go-qrcode"
)

/*
Email Service with Template Support

This service provides two email sending functions:

1. SendPaymentSuccessEmailWithQR (existing) - Simple email with QR attachment
2. SendRegistrationSuccessEmail (new) - HTML template-based email with CID-referenced images

New Template Email Features:
- Uses HTML templates (templates/email_template_en.html, templates/email_template_vi.html)
- Embeds QR code and logos as inline attachments with CID references
- Better email deliverability compared to base64 embedded images
- Loads event information from environment variables (EVENT_NAME, EVENT_DATE, EVENT_VENUE)
- Supports both English and Vietnamese languages

Usage Example:
	templateData := TemplateData{
		FullName:        "John Doe",
		PhoneNumber:     "+1234567890",
		RegistrationFee: "$500",
	}

	err := SendRegistrationSuccessEmail(
		"user@example.com", // To email
		"John Doe",         // To name
		"reg123",          // Registration ID
		"en",              // Language ("en" or "vi")
		templateData,      // Template data
		config,           // App config
	)

Environment Variables Required:
- EVENT_NAME: Name of the event
- EVENT_DATE: Date of the event
- EVENT_VENUE: Venue of the event
- SEND_GRIP_API_KEY: SendGrid API key
- SEND_GRIP_SENDER_NAME: Sender name
- SEND_GRIP_SENDER_EMAIL: Sender email

Logo files should be placed in templates/ directory:
- templates/logo_1.png
- templates/logo_2.png
- templates/logo_3.png

Template Data Structure:
- FullName: Participant's full name
- PhoneNumber: Participant's phone number
- RegistrationFee: Registration fee amount
- EventName: Event name (loaded from config)
- EventDate: Event date (loaded from config)
- EventVenue: Event venue (loaded from config)

Images (QR code and logos) are attached as inline attachments with CID references.
*/

type TemplateData struct {
	FullName        string
	PhoneNumber     string
	RegistrationFee string
	EventName       string
	EventDate       string
	EventVenue      string
}

// AddInlineAttachment adds an inline attachment with CID to the email
func AddInlineAttachment(message *mail.SGMailV3, filePath string, cid string, contentType string) error {
	if filePath == "" {
		return nil // Skip if no file path provided
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Warning: Could not load file from %s: %v", filePath, err)
		return nil // Return nil to allow email to proceed without this attachment
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Warning: Could not read file %s: %v", filePath, err)
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString(fileBytes)
	attachment := mail.NewAttachment()
	attachment.SetContent(encoded)
	attachment.SetType(contentType)
	attachment.SetFilename(cid + ".png")
	attachment.SetDisposition("inline")
	attachment.SetContentID(cid)

	message.AddAttachment(attachment)
	return nil
}

// SendRegistrationSuccessEmail sends email using HTML templates with CID-referenced images
func SendRegistrationSuccessEmail(
	toEmail, toName, registerID, language string,
	templateData TemplateData,
	config *config.Config,
) error {
	from := mail.NewEmail(config.SendGrip.SenderName, config.SendGrip.SenderEmail)
	to := mail.NewEmail(toName, toEmail)

	var subject string
	var templatePath string

	// Choose template and subject based on language
	if language == "vi" {
		subject = "🎉 Xác nhận đăng ký thành công - ASHNO 2025"
		templatePath = "templates/email_template_vi.html"
	} else {
		subject = "🎉 Registration Confirmation - ASHNO 2025"
		templatePath = "templates/email_template_en.html"
	}

	// Load event info from config
	templateData.EventName = config.Event.Name
	templateData.EventDate = config.Event.Date
	templateData.EventVenue = config.Event.Venue

	// Parse and execute template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, templateData)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := htmlBuffer.String()

	// Create email message
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	// Generate and add QR code as inline attachment
	qrURL := fmt.Sprintf("%s/%s", "https://checkout-ashno2025.vercel.app", registerID)
	png, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Add QR code as inline attachment
	encoded := base64.StdEncoding.EncodeToString(png)
	qrAttachment := mail.NewAttachment()
	qrAttachment.SetContent(encoded)
	qrAttachment.SetType("image/png")
	qrAttachment.SetFilename("qrcode.png")
	qrAttachment.SetDisposition("inline")
	qrAttachment.SetContentID("qrcode")
	message.AddAttachment(qrAttachment)

	// Add logo attachments
	AddInlineAttachment(message, "templates/logo_1.png", "logo1", "image/png")
	AddInlineAttachment(message, "templates/logo_2.png", "logo2", "image/png")
	AddInlineAttachment(message, "templates/logo_3.png", "logo3", "image/png")

	// Send email
	client := sendgrid.NewSendClient(config.SendGrip.ApiKey)
	_, err = client.Send(message)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

// Generate QR and send email with attachment
func SendPaymentSuccessEmailWithQR(
	toEmail, toName, registerID string,
	config *config.Config,
) error {
	from := mail.NewEmail(config.SendGrip.SenderName, config.SendGrip.SenderEmail)
	subject := "🎉 Payment Confirmation - QR Ticket Attached"
	to := mail.NewEmail(toName, toEmail)

	// Generate QR code
	qrURL := fmt.Sprintf("%s/%s", "https://checkout-ashno2025.vercel.app", registerID)
	png, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Create attachment
	encoded := base64.StdEncoding.EncodeToString(png)
	attachment := mail.NewAttachment()
	attachment.SetContent(encoded)
	attachment.SetType("image/png")
	attachment.SetFilename("qr_code.png")
	attachment.SetDisposition("attachment")

	// Email content
	htmlContent := fmt.Sprintf(`
		Hi %s,<br><br>
		Your registration was successful!<br>
		Scan the attached QR code at the event check-in.<br><br>
		Thanks,<br>
		ASHNO 2025
	`, toName)

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	message.AddAttachment(attachment)

	client := sendgrid.NewSendClient(config.SendGrip.ApiKey)
	_, err = client.Send(message)
	if err != nil {
		log.Println(err)
	}
	return err
}

// SendRegistrationEmailWithTemplate is a convenience function for easy usage
func SendRegistrationEmailWithTemplate(
	toEmail, toName, registerID, language string,
	fullName, phoneNumber, registrationFee string,
	config *config.Config,
) error {
	// Prepare template data
	templateData := TemplateData{
		FullName:        fullName,
		PhoneNumber:     phoneNumber,
		RegistrationFee: registrationFee,
		// EventName, EventDate, EventVenue will be loaded from config in SendRegistrationSuccessEmail
	}

	return SendRegistrationSuccessEmail(
		toEmail,
		toName,
		registerID,
		language,
		templateData,
		config,
	)
}
