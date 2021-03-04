package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
)

type Application struct {
	Template *template.Template
	Redis    *redis.Client
}

func (a *Application) Index(w http.ResponseWriter, r *http.Request) {
	keys, err := a.Redis.Keys(r.Context(), "*").Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Template.Execute(w, keys)
}

type report struct {
	CreatedAt      *time.Time `json:"created_at"`
	Amount         *float64   `json:"amount"`
	UserId         *string    `json:"user_id"`
	AgentId        *string    `json:"agent_id"`
	Percent        *float64   `json:"percent"`
	PaymentId      *uuid.UUID `json:"payment_id"`
	Number         *int       `json:"number"`
	UserLastname   *string    `json:"user_lastname"`
	UserFirstname  *string    `json:"user_firstname"`
	AgentLastname  *string    `json:"agent_lastname"`
	AgentFirstname *string    `json:"agent_firstname"`
	Birthday       *time.Time `json:"birthday"`
}

func (r report) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (r report) Valid() bool {
	return r.CreatedAt != nil &&
		r.Amount != nil &&
		r.UserId != nil &&
		r.AgentId != nil &&
		r.Percent != nil &&
		r.PaymentId != nil &&
		r.Number != nil &&
		r.UserLastname != nil &&
		r.UserFirstname != nil &&
		r.AgentLastname != nil &&
		r.AgentFirstname != nil &&
		r.Birthday != nil
}

func (a *Application) Report(w http.ResponseWriter, r *http.Request) {
	v, err := a.Redis.Get(r.Context(), mux.Vars(r)["id"]).Result()
	if err == nil {
		var j report
		err = json.Unmarshal([]byte(v), &j)
		if err == nil && j.Valid() {
			a.pdf(w, j)
			return
		}
	}
	http.NotFound(w, r)
}

func (a *Application) pdf(w http.ResponseWriter, j report) error {
	pdf := gofpdf.New("P", "mm", "A4", "fonts")
	pdf.AddUTF8Font("DejaVu", "", "DejaVuSans.ttf")
	pdf.AddPage()
	pdf.SetFont("DejaVu", "", 16)
	pdf.Write(8, "Отчет")
	pdf.Ln(8)
	pdf.Writef(8, "Учреждение - %s %s", *j.AgentFirstname, *j.AgentLastname)
	pdf.Ln(8)
	t := time.Now()
	pdf.Writef(8, "Отчетный период - c %d.%02d.%02d по %d.%02d.%02d",
		j.CreatedAt.Year(), j.CreatedAt.Month(), j.CreatedAt.Day(),
		t.Year(), t.Month(), t.Day())
	pdf.Ln(8)
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 14)
	pdf.Cell(40, 7, "user_id:")
	pdf.Write(7, *j.UserId)
	pdf.Ln(7)
	pdf.Cell(40, 7, "agent_id:")
	pdf.Write(7, *j.AgentId)
	pdf.Ln(7)
	pdf.Cell(40, 7, "percent:")
	pdf.Writef(7, "%.2f%%", *j.Percent)
	pdf.Ln(7)
	pdf.Cell(40, 7, "payment_id:")
	pdf.Writef(7, "%v", *j.PaymentId)
	pdf.Ln(7)
	pdf.Cell(40, 7, "number:")
	pdf.Writef(7, "%d", *j.Number)
	pdf.Ln(7)
	pdf.Cell(40, 7, "lastname:")
	pdf.Writef(7, *j.UserLastname)
	pdf.Ln(7)
	pdf.Cell(40, 7, "firstname:")
	pdf.Writef(7, *j.UserFirstname)
	pdf.Ln(7)
	pdf.Cell(40, 7, "birthday:")
	pdf.Writef(7, "%d.%02d.%02d", j.Birthday.Year(), j.Birthday.Month(), j.Birthday.Day())
	pdf.Ln(7)
	w.Header().Set("Content-Type", "application/pdf")
	return pdf.Output(w)
}

func main() {
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatal(err)
	}
	a := Application{
		Redis: redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		}),
		Template: t,
	}
	h := mux.NewRouter()
	h.HandleFunc("/", a.Index).Methods(http.MethodGet, http.MethodHead)
	h.HandleFunc("/index.html", a.Index).Methods(http.MethodGet, http.MethodHead)
	h.HandleFunc("/{id}.pdf", a.Report).Methods(http.MethodGet)
	err = http.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}
