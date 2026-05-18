package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	catalogmodel "github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
	"github.com/setorin/setorin/backend/internal/apps/order/application/dto"
	orderports "github.com/setorin/setorin/backend/internal/apps/order/application/ports"
	ordermodel "github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newOrderUC(orderRepo *MockOrderRepository, matRepo *MockMaterialRepository, userRepo *MockUserRepository) *UserOrderUseCase {
	return NewUserOrderUseCase(orderRepo, matRepo, userRepo, nil) // mail=nil for unit tests
}

func activeMaterial(slug string) *catalogmodel.MaterialWithPrice {
	price := int64(5500)
	return &catalogmodel.MaterialWithPrice{
		Material: catalogmodel.Material{
			ID:       uuid.New(),
			Slug:     slug,
			Name:     "Plastik",
			Unit:     "kg",
			IsActive: true,
		},
		CurrentPrice: &price,
	}
}

func receivedOrder(sellerID, materialID uuid.UUID) *ordermodel.Order {
	return &ordermodel.Order{
		ID:                uuid.New(),
		OrderCode:         "ECC-00001",
		UserID:            sellerID,
		MaterialID:        materialID,
		EstimatedWeightKg: "3.500",
		UnitPriceAtOrder:  5500,
		EstimatedPayout:   19250,
		Method:            ordermodel.MethodPickup,
		Status:            ordermodel.StatusReceived,
		AddressText:       "Jl. Test No. 1, Jakarta",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func sellerUser(id uuid.UUID) *usermodel.User {
	return &usermodel.User{
		ID:       id,
		Email:    "seller@example.com",
		FullName: "Seller User",
		Status:   usermodel.StatusActive,
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	created := receivedOrder(sellerID, mat.ID)

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "plastic").Return(mat, nil)
	orderRepo.On("NextOrderCode", mock.Anything).Return("ECC-00001", nil)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(created, nil)
	userRepo.On("FindByID", mock.Anything, sellerID).Return(sellerUser(sellerID), nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.Create(context.Background(), sellerID, dto.CreateOrderRequest{
		MaterialSlug:      "plastic",
		EstimatedWeightKg: "3.500",
		Method:            "pickup",
		AddressText:       "Jl. Test No. 1, Jakarta",
	})

	require.NoError(t, err)
	assert.Equal(t, "ECC-00001", o.OrderCode)
	assert.Equal(t, ordermodel.StatusReceived, o.Status)
}

func TestCreate_MaterialNotFound(t *testing.T) {
	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "unknown").Return(nil, apperr.New(apperr.CodeNotFound, "not found"))

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Create(context.Background(), uuid.New(), dto.CreateOrderRequest{
		MaterialSlug:      "unknown",
		EstimatedWeightKg: "1.000",
		Method:            "pickup",
		AddressText:       "Jl. Test No. 1, Jakarta",
	})

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeNotFound, ae.Code)
}

func TestCreate_MaterialInactive(t *testing.T) {
	mat := activeMaterial("plastic")
	mat.IsActive = false

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "plastic").Return(mat, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Create(context.Background(), uuid.New(), makeCreateReq("plastic", "1.000"))

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeValidation, ae.Code)
}

func TestCreate_NoPrice_ReturnsValidationError(t *testing.T) {
	mat := activeMaterial("plastic")
	mat.CurrentPrice = nil // no price set

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "plastic").Return(mat, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Create(context.Background(), uuid.New(), makeCreateReq("plastic", "1.000"))

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeValidation, ae.Code)
}

func TestCreate_InvalidWeight_ReturnsValidationError(t *testing.T) {
	mat := activeMaterial("plastic")

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "plastic").Return(mat, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Create(context.Background(), uuid.New(), makeCreateReq("plastic", "abc"))

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeValidation, ae.Code)
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func TestCancel_Success(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)
	cancelled := *order
	cancelled.Status = ordermodel.StatusCancelled

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)
	orderRepo.On("Transition", mock.Anything, order.ID, mock.MatchedBy(func(p orderports.TransitionParams) bool {
		return p.NewStatus == ordermodel.StatusCancelled && p.ExpectedStatus == ordermodel.StatusReceived
	})).Return(&cancelled, nil)
	userRepo.On("FindByID", mock.Anything, sellerID).Return(sellerUser(sellerID), nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.Cancel(context.Background(), sellerID, "ECC-00001", "no longer needed")

	require.NoError(t, err)
	assert.Equal(t, ordermodel.StatusCancelled, o.Status)
}

func TestCancel_WrongOwner_Forbidden(t *testing.T) {
	sellerID := uuid.New()
	otherID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Cancel(context.Background(), otherID, "ECC-00001", "reason")

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeForbidden, ae.Code)
}

func TestCancel_WrongStatus_Conflict(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)
	order.Status = ordermodel.StatusAccepted // already accepted — can't cancel via user flow

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.Cancel(context.Background(), sellerID, "ECC-00001", "reason")

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeConflict, ae.Code)
}

// ─── ConfirmCash ─────────────────────────────────────────────────────────────

func TestConfirmCash_Success(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)
	order.Status = ordermodel.StatusCashHandover

	finalPayout := int64(19250)
	done := *order
	done.Status = ordermodel.StatusDone
	done.FinalPayout = &finalPayout

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)
	orderRepo.On("Transition", mock.Anything, order.ID, mock.MatchedBy(func(p orderports.TransitionParams) bool {
		return p.NewStatus == ordermodel.StatusDone && p.ExpectedStatus == ordermodel.StatusCashHandover
	})).Return(&done, nil)
	userRepo.On("FindByID", mock.Anything, sellerID).Return(sellerUser(sellerID), nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.ConfirmCash(context.Background(), sellerID, "ECC-00001")

	require.NoError(t, err)
	assert.Equal(t, ordermodel.StatusDone, o.Status)
}

func TestConfirmCash_WrongOwner_Forbidden(t *testing.T) {
	sellerID := uuid.New()
	otherID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)
	order.Status = ordermodel.StatusCashHandover

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.ConfirmCash(context.Background(), otherID, "ECC-00001")

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeForbidden, ae.Code)
}

func TestConfirmCash_WrongStatus_Conflict(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)
	order.Status = ordermodel.StatusQuality // not at cash_handover yet

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.ConfirmCash(context.Background(), sellerID, "ECC-00001")

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeConflict, ae.Code)
}

// ─── GetByCode ───────────────────────────────────────────────────────────────

func TestGetByCode_OwnerCanView(t *testing.T) {
	sellerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.GetByCode(context.Background(), sellerID, "ECC-00001", false)

	require.NoError(t, err)
	assert.Equal(t, order.ID, o.ID)
}

func TestGetByCode_Stranger_Forbidden(t *testing.T) {
	sellerID := uuid.New()
	strangerID := uuid.New()
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	_, err := uc.GetByCode(context.Background(), strangerID, "ECC-00001", false)

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeForbidden, ae.Code)
}

func TestGetByCode_AdminCanView(t *testing.T) {
	sellerID := uuid.New()
	adminID := uuid.New() // different from seller
	mat := activeMaterial("plastic")
	order := receivedOrder(sellerID, mat.ID)

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	orderRepo.On("FindByCode", mock.Anything, "ECC-00001").Return(order, nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.GetByCode(context.Background(), adminID, "ECC-00001", true) // isAdmin=true

	require.NoError(t, err)
	assert.Equal(t, order.ID, o.ID)
}

// ─── EstimatedPayout calculation ─────────────────────────────────────────────

func TestCreate_PayoutCalculation(t *testing.T) {
	sellerID := uuid.New()
	price := int64(18500) // aluminium
	mat := &catalogmodel.MaterialWithPrice{
		Material: catalogmodel.Material{
			ID:       uuid.New(),
			Slug:     "aluminum",
			Name:     "Aluminium",
			Unit:     "kg",
			IsActive: true,
		},
		CurrentPrice: &price,
	}

	// 2.5 kg × 18500/kg = 46250
	expectedPayout := int64(46250)

	var capturedOrder *ordermodel.Order
	returned := &ordermodel.Order{
		ID:              uuid.New(),
		OrderCode:       "ECC-00002",
		UserID:          sellerID,
		MaterialID:      mat.ID,
		EstimatedPayout: expectedPayout,
		Status:          ordermodel.StatusReceived,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	orderRepo := &MockOrderRepository{}
	matRepo := &MockMaterialRepository{}
	userRepo := &MockUserRepository{}

	matRepo.On("FindBySlug", mock.Anything, "aluminum").Return(mat, nil)
	orderRepo.On("NextOrderCode", mock.Anything).Return("ECC-00002", nil)
	orderRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *ordermodel.Order) bool {
		capturedOrder = o
		return true
	})).Return(returned, nil)
	userRepo.On("FindByID", mock.Anything, sellerID).Return(sellerUser(sellerID), nil)

	uc := newOrderUC(orderRepo, matRepo, userRepo)
	o, err := uc.Create(context.Background(), sellerID, makeCreateReqFull("aluminum", "2.500", "pickup", "Jl. Test No. 1, Jakarta"))

	require.NoError(t, err)
	require.NotNil(t, capturedOrder)
	assert.Equal(t, expectedPayout, capturedOrder.EstimatedPayout)
	assert.Equal(t, "2.500", capturedOrder.EstimatedWeightKg)
	assert.Equal(t, price, capturedOrder.UnitPriceAtOrder)
	assert.Equal(t, expectedPayout, o.EstimatedPayout)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func makeCreateReq(slug, weight string) dto.CreateOrderRequest {
	return makeCreateReqFull(slug, weight, "pickup", "Jl. Test No. 1, Jakarta Selatan")
}

func makeCreateReqFull(slug, weight, method, address string) dto.CreateOrderRequest {
	return dto.CreateOrderRequest{
		MaterialSlug:      slug,
		EstimatedWeightKg: weight,
		Method:            method,
		AddressText:       address,
	}
}
