package domain

import (
	"errors"
	"testing"
)

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		from, to Status
		want     bool
	}{
		{StatusCreated, StatusProcessed, true},
		{StatusCreated, StatusDenied, true},
		{StatusCreated, StatusComplected, false},
		{StatusCreated, StatusCreated, false},
		{StatusProcessed, StatusComplected, true},
		{StatusProcessed, StatusCreated, false},
		{StatusComplected, StatusDelivered, true},
		{StatusComplected, StatusDenied, true},
		{StatusDelivered, StatusDenied, false},
		{StatusDenied, StatusProcessed, false},
	}

	for _, tt := range tests {
		if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
			t.Errorf("%s -> %s = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestStatusValid(t *testing.T) {
	if !StatusDenied.Valid() {
		t.Error("denied must be valid")
	}
	if Status(42).Valid() {
		t.Error("42 is not a status")
	}
}

// Ошибки checkout несут детали позиции, но снаружи ловятся по sentinel-ам.
func TestErrorsUnwrapToSentinels(t *testing.T) {
	stock := &StockError{ProductUUID: "u", Requested: 3, Available: 1}
	if !errors.Is(stock, ErrInsufficientStock) {
		t.Error("StockError must unwrap to ErrInsufficientStock")
	}

	transition := &TransitionError{From: StatusDelivered, To: StatusCreated}
	if !errors.Is(transition, ErrInvalidStatusTransition) {
		t.Error("TransitionError must unwrap to ErrInvalidStatusTransition")
	}
}
