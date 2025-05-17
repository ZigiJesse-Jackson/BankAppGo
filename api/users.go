package api

import (
	db "BankAppGo/db/sqlc"
	"BankAppGo/db/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username  string `form:"username" binding:"required,alphanum"`
	Password  string `form:"password" binding:"required,alphanum,min=6"`
	Full_name string `form:"fullname" binding:"required"`
	Email     string `form:"email" binding:"required,email"`
}

type createUserResponse struct {
	Username          string
	Full_name         string
	Email             string
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	// binds HTTP request to json which is then extracted to struct
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashed_pwd, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// prepare create account params
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashed_pwd,
		Email:          req.Email,
		FullName:       req.Full_name,
	}

	// creates account
	user, err := server.store.CreateUser(ctx, arg)

	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := createUserResponse{
		Username:          user.Username,
		Full_name:         user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)

}
