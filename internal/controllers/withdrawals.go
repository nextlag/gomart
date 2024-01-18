package controllers

//
// import (
// 	"errors"
// 	"net/http"
//
// 	"github.com/nextlag/gomart/internal/mw/auth"
// 	"github.com/nextlag/gomart/internal/mw/logger"
// )
//
// type Users struct {
// }
//
// func (c Controller) Withdrawals(w http.ResponseWriter, r *http.Request) {
// 	user := r.Context().Value(auth.LoginKey).(string)
//
// 	var spendRequest getSpendBonusRequest
//
// 	if err := ctx.ShouldBindJSON(&spendRequest); err != nil {
// 		logger.ErrorLogger("Wrong request: ", err)
// 		ctx.JSON(http.StatusBadRequest, newErrorMessage("Wrong request"))
// 		return
// 	}
//
// 	err := h.s.SpendBonuses(ctx, login, spendRequest.Order, spendRequest.Sum)
// 	switch {
// 	case errors.Is(err, psql.ErrNotEnoughBalance):
// 		ctx.AbortWithStatusJSON(http.StatusPaymentRequired, newErrorMessage("Not enough balance"))
// 		return
// 	case errors.Is(err, validitycheck.ErrWrongOrderNum):
// 		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, newErrorMessage("Wrong order number"))
// 		return
// 	case errors.Is(err, psql.ErrAlreadyLoadedOrder) || errors.Is(err, psql.ErrYouAlreadyLoadedOrder):
// 		ctx.AbortWithStatusJSON(http.StatusConflict, newErrorMessage("Order is already loaded"))
// 		return
// 	case err != nil:
// 		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, newMessage("Bonuses successfully spent"))
//
// 	w.Write([]byte(user))
// }
