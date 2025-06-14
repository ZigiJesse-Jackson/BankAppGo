package api

import (
	db "BankAppGo/db/sqlc"
	"BankAppGo/token"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `form:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `form:"to_account_id" binding:"required,min=1"`
	Amount        int64  `form:"amount" binding:"required,gt=0"`
	Currency      string `form:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// check currency transfer will occur in
	account, valid := server.validate_account(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	if account.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if _, valid = server.validate_account(ctx, req.ToAccountID, req.Currency); !valid {
		return
	}

	// prepare create transfer params
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	// creates transfer
	transfer, err := server.store.TransferTx(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)

}

func (server *Server) validate_account(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, int64(accountID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse((err)))
			return db.Account{}, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return db.Account{}, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return db.Account{}, false
	}

	return account, true

}
