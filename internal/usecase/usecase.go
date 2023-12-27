package usecase

import (
	"github.com/nextlag/gomart/internal/entity"
)

type Entity interface {
	GetEntityRequest() *entity.Request
}

type UseCase struct {
	entity Entity // interface Entity
	req    entity.Request
}

func (uc *UseCase) GetEntityRequest() *entity.Request {
	return &uc.req
}

func New(e Entity) *UseCase {
	r := UseCase{}.req
	return &UseCase{e, r}
}

func (uc *UseCase) DoRequest() {
	uc.entity.GetEntityRequest()
}

func NewRequest(login, order, status, uploadedAt string, bonusesWithdrawn, accrual *float32, id int, user, pass string, time string) Entity {
	return &UseCase{
		req: entity.Request{
			Order: entity.Order{
				Login:            login,
				Order:            order,
				Status:           status,
				UploadedAt:       uploadedAt,
				BonusesWithdrawn: bonusesWithdrawn,
				Accrual:          accrual,
			},
			User: entity.User{
				ID:       id,
				Username: user,
				Password: pass,
			},
			Points: entity.Points{
				Order:   order,
				Status:  status,
				Accrual: accrual,
			},
			OWSB: entity.OWSB{
				Order:            order,
				Time:             time,
				BonusesWithdrawn: bonusesWithdrawn,
			},
		},
	}
}
