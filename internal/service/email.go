package service

import (
	"ashno-onepay/internal/config"
	_ "bytes"
	"encoding/base64"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/skip2/go-qrcode"
	"log"
)

// r.config.Server.Host+":"+r.config.Server.Port,
//
//	r.config.SendGrip.ApiKey,
//
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
