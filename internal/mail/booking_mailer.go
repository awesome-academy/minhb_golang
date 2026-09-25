package mail

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
	"gopkg.in/gomail.v2"

	"cinema-booking/internal/models"
	"cinema-booking/internal/utils"
)

const (
	bookingCreatedTemplate = "booking-created.html"
	bookingPaidTemplate    = "booking-paid.html"
	mailTimeLayout         = "Mon 02 Jan 2006 15:04"
	imageTimeLayout        = "2006-01-02-15h04"
	qrSize                 = 256
)

type Options struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type BookingMailer struct {
	dialer    *gomail.Dialer
	from      string
	templates *template.Template
}

func NewBookingMailer(opts Options, templates *template.Template) *BookingMailer {
	return &BookingMailer{
		dialer:    gomail.NewDialer(opts.Host, opts.Port, opts.Username, opts.Password),
		from:      opts.From,
		templates: templates,
	}
}

func (m *BookingMailer) SendBookingCreated(ctx context.Context, booking *models.Booking) error {
	msg, err := m.newMessage(booking, "Your booking · "+booking.Showtime.Movie.Title, bookingCreatedTemplate)
	if err != nil {
		return err
	}
	return m.send(ctx, msg)
}

func (m *BookingMailer) SendBookingPaid(ctx context.Context, booking *models.Booking) error {
	msg, err := m.newMessage(booking, "Your tickets · "+booking.Showtime.Movie.Title, bookingPaidTemplate)
	if err != nil {
		return err
	}
	if err := embedTicketQRCodes(msg, booking); err != nil {
		return err
	}
	return m.send(ctx, msg)
}

func (m *BookingMailer) newMessage(booking *models.Booking, subject, templateName string) (*gomail.Message, error) {
	var body bytes.Buffer
	if err := m.templates.ExecuteTemplate(&body, templateName, newBookingView(booking)); err != nil {
		return nil, err
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetAddressHeader("To", booking.User.Email, booking.User.FullName)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())
	return msg, nil
}

func (m *BookingMailer) send(ctx context.Context, msg *gomail.Message) error {
	done := make(chan error, 1)
	go func() { done <- m.dialer.DialAndSend(msg) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func embedTicketQRCodes(msg *gomail.Message, booking *models.Booking) error {
	for _, ticket := range booking.Tickets {
		png, err := qrcode.Encode(ticket.QRCode, qrcode.Medium, qrSize)
		if err != nil {
			return err
		}
		msg.Embed(ticketImage(booking.Showtime, ticket), gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(png)
			return err
		}))
	}
	return nil
}

type bookingView struct {
	FullName   string
	Code       string
	MovieTitle string
	Format     string
	Theater    string
	Address    string
	Room       string
	StartsAt   string
	ExpiresAt  string
	PaidAt     string
	Total      string
	Currency   string
	Tickets    []ticketView
}

type ticketView struct {
	Seat  string
	Price string
	Image string
}

func newBookingView(booking *models.Booking) bookingView {
	showtime := booking.Showtime
	tickets := make([]ticketView, 0, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		tickets = append(tickets, ticketView{
			Seat:  seatLabel(ticket.Seat),
			Price: ticket.Price.StringFixed(2),
			Image: ticketImage(showtime, ticket),
		})
	}
	return bookingView{
		FullName:   booking.User.FullName,
		Code:       booking.Code,
		MovieTitle: showtime.Movie.Title,
		Format:     string(showtime.Format),
		Theater:    showtime.Room.Theater.Name,
		Address:    showtime.Room.Theater.Address,
		Room:       showtime.Room.Name,
		StartsAt:   utils.FormatVN(showtime.StartsAt, mailTimeLayout),
		ExpiresAt:  formatOptional(booking.ExpiresAt),
		PaidAt:     formatOptional(booking.ConfirmedAt),
		Total:      booking.Subtotal.Sub(booking.DiscountAmount).StringFixed(2),
		Currency:   booking.Currency,
		Tickets:    tickets,
	}
}

func formatOptional(t *time.Time) string {
	if t == nil {
		return ""
	}
	return utils.FormatVN(*t, mailTimeLayout)
}

func ticketImage(showtime models.Showtime, ticket models.Ticket) string {
	return utils.Slugify(showtime.Movie.Title) + "_" + utils.FormatVN(showtime.StartsAt, imageTimeLayout) + "_" + seatLabel(ticket.Seat) + ".png"
}

func seatLabel(seat models.Seat) string {
	return seat.RowLabel + strconv.Itoa(int(seat.SeatNumber))
}
