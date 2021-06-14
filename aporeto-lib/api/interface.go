package api

import (
	"context"

	"go.aporeto.io/manipulate"
)

// CreatorDeleter interface defines APIs to create and delete objects.
type CreatorDeleter interface {

	// Create creates an object.
	Create(ctx context.Context, m manipulate.Manipulator) error

	// Delete deletes an object.
	Delete(ctx context.Context, m manipulate.Manipulator) error
}

// Disabler interface defines APIs to disable an object.
type Disabler interface {
	// Disable disables an object.
	Disable(ctx context.Context, m manipulate.Manipulator) error
}

// CreatorDeleterDisabler interface composes the CreatorDeleter and Disabler interfaces.
type CreatorDeleterDisabler interface {
	CreatorDeleter
	Disabler
}
