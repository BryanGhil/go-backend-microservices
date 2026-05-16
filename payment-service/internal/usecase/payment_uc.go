package usecase

import (
	"context"
	"ecommerce/payment-service/internal/domain"
)

type paymentUC struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(r domain.PaymentRepository) domain.PaymentUseCase {
	return &paymentUC{repo: r}
}

func (u *paymentUC) ProcessPayment(ctx context.Context, event domain.SagaEvent) (bool, error) {
	
	// --- STRIPE / MIDTRANS INTEGRATION GOES HERE ---
	// Example:
	// params := &stripe.ChargeParams{ Amount: stripe.Int64(int64(event.Amount * 100)), Currency: stripe.String(string(stripe.CurrencyUSD)) }
	// charge, err := charge.New(params)
	// success := err == nil
	// -----------------------------------------------

	// MOCK LOGIC FOR SAGA TESTING:
	// If amount is greater than 1000, simulate a DECLINED card (Insufficient Funds)
	success := true
	status := "SUCCESS"
	
	if event.TotalAmount > 1000.00 {
		success = false
		status = "DECLINED"
	}

	// Save to our ledger
	payment := &domain.Payment{
		CorrelationID: event.CorrelationID, // FIX: Types now match perfectly!
		Amount:        event.TotalAmount,
		Status:        status,
	}
	u.repo.SaveTransaction(ctx, payment)

	return success, nil
}

func (u *paymentUC) GetStatus(ctx context.Context, correlationID string) (string, error) {
	return u.repo.GetStatusByCorrelationID(ctx, correlationID)
}