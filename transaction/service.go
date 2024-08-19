package transaction

import (
	"crowdfunding/campaign"
	"crowdfunding/payment"
	"errors"
	"time"
)

type service struct {
	repository         Repository
	campaignRepository campaign.Repository
	paymentService     payment.Service
}

type Service interface {
	GetTransactionByCampaignID(input GetCampaignTransactionsInput) ([]Transactions, error)
	GetTransactionByUserID(userID int) ([]Transactions, error)
	CreateTransaction(input CreateTransactionInput) (Transactions, error)
}

func NewService(repository Repository, campaignRepository campaign.Repository, paymentService payment.Service) *service {
	return &service{repository, campaignRepository, paymentService}
}

func (s *service) GetTransactionByCampaignID(input GetCampaignTransactionsInput) ([]Transactions, error) {

	//get campaign by id
	//check campaign.userid != user.id yang melakukan request

	campaign, err := s.campaignRepository.FindByID(input.ID)

	if err != nil {
		return []Transactions{}, err
	}

	if campaign.UserID != input.User.ID {
		return []Transactions{}, errors.New("Unauthorized")
	}

	transactions, err := s.repository.GetByCampaignID(input.ID)
	if err != nil {
		return transactions, err
	}
	return transactions, nil
}

func (s *service) GetTransactionByUserID(userID int) ([]Transactions, error) {
	transactions, err := s.repository.GetByUserID(userID)

	if err != nil {
		return transactions, err
	}

	return transactions, nil
}

func (s *service) CreateTransaction(input CreateTransactionInput) (Transactions, error) {
	transaction := Transactions{}
	transaction.CampaignID = input.CampaignID
	transaction.Amount = input.Amount
	transaction.UserID = input.User.ID
	transaction.Status = "pending"
	transaction.Code = "ORDER-" + time.Now().Format("20060102150405")

	newTransaction, err := s.repository.Save(transaction)
	if err != nil {
		return newTransaction, err
	}
	paymentTransaction := payment.Transaction{
		ID:     newTransaction.ID,
		Amount: newTransaction.Amount,
	}
	paymentURL, err := s.paymentService.GetPaymentURL(paymentTransaction, input.User)
	{
		if err != nil {
			return newTransaction, nil
		}
		newTransaction.PaymentURL = paymentURL

		newTransaction, err = s.repository.Update(newTransaction)

		if err != nil {
			return newTransaction, err
		}

		return newTransaction, nil
	}
}
