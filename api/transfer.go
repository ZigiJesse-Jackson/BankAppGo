package api

import (
	db "BankAppGo/db/sqlc"
	"database/sql"
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

	// check currency transfer will occur in
	if !server.validate_account(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.validate_account(ctx, req.ToAccountID, req.Currency) {
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

func (server *Server) validate_account(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, int64(accountID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse((err)))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true

}
