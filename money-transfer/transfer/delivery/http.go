package delivery

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	middleware "money-transfer/transfer/delivery/middleware"

	"github.com/rs/zerolog/log"

	"money-transfer/domain"

	"github.com/go-chi/chi/v5"
)

type TransferHanlder struct {
	usecase domain.Transfer
}

func NewTransactionHandler(c *domain.Config, router *chi.Mux, tu domain.Transfer) error {
	handler := &TransferHanlder{
		usecase: tu,
	}

	fs := http.FileServer(http.Dir("static/css"))
	router.Handle("/static/*", http.StripPrefix("/static/css", fs))

	tmpl, err := template.ParseFiles("./static/html/login.html", "./static/html/main.html")
	if err != nil {
		return err
	}

	router.Get("/login", handler.LoginPage(tmpl))

	m := middleware.New(c)

	router.With(m.CheckAuthMiddleware).Get("/main", handler.IndexPage(tmpl))
	router.With(m.CheckAuthMiddleware).Get("/accounts", handler.AccountsInfo)
	router.With(m.CheckAuthMiddleware).Post("/accounts", handler.CreateAccount)
	router.With(m.CheckAuthMiddleware).Post("/transaction", handler.SendMoney)
	router.With(m.CheckAuthMiddleware).Post("/increment", handler.TopUpAccount)
	return nil
}

func (th *TransferHanlder) IndexPage(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := tmpl.ExecuteTemplate(w, "main", nil); err != nil {

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (th *TransferHanlder) LoginPage(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := tmpl.ExecuteTemplate(w, "login", nil); err != nil {

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (th *TransferHanlder) AccountsInfo(w http.ResponseWriter, r *http.Request) {

	u, ok := r.Context().Value(middleware.CtxKeyUser).(*domain.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accounts, err := th.usecase.GetAccounts(r.Context(), u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn().Err(err).Msg("error with account")
		return
	}

	reply, err := json.Marshal(accounts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warn().Err(err).Msg("error with marshal")
		return
	}

	w.Write(reply)

}

func (th *TransferHanlder) CreateAccount(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(middleware.CtxKeyUser).(*domain.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	IIN := r.FormValue("IIN")

	if u.IIN != IIN {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("incorrect iin"))
		return
	}

	account := domain.Account{
		OwnerID:    u.ID,
		IIN:        IIN,
		Amount:     0,
		Registered: time.Now(),
	}

	err := th.usecase.CreateAccount(r.Context(), &account)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Created"))
}

func (th *TransferHanlder) SendMoney(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(middleware.CtxKeyUser).(*domain.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	from := r.FormValue("senderID")
	to := r.FormValue("recieverID")
	value := r.FormValue("amount")

	sender, err := strconv.ParseInt(from, 10, 64)
	receiver, err2 := strconv.ParseInt(to, 10, 64)
	amount, err3 := strconv.ParseInt(value, 10, 64)

	if err != nil || err2 != nil || err3 != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = th.usecase.CreateTransaction(r.Context(), u.ID, sender, receiver, amount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

func (th *TransferHanlder) TopUpAccount(w http.ResponseWriter, r *http.Request) {
	acc := r.FormValue("accountID")
	value := r.FormValue("amount")

	accountID, err := strconv.ParseInt(acc, 10, 64)
	amount, err2 := strconv.ParseInt(value, 10, 64)
	if err != nil || err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = th.usecase.ChangeAccountSum(r.Context(), accountID, amount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
