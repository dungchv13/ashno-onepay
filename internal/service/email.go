package service

import (
	_ "bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/skip2/go-qrcode"
)

// Generate QR and send email with attachment
func SendPaymentSuccessEmailWithQR(
	toEmail, toName, registerID string,
	amount int64,
	currency, hostURL string, // e.g., "https://yourdomain.com"
) error {

	from := mail.NewEmail("Your Event Team", "noreply@yourevent.com")
	subject := "🎉 Payment Confirmation - QR Ticket Attached"
	to := mail.NewEmail(toName, toEmail)

	// Generate QR code
	qrURL := fmt.Sprintf("%s/register/%s/registration-info", hostURL, registerID)
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
		<b>Amount:</b> %d %s<br><br>
		Scan the attached QR code at the event check-in.<br><br>
		Thanks,<br>
		Event Registration Team
	`, toName, amount, currency)

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	message.AddAttachment(attachment)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err = client.Send(message)

	return err
}
