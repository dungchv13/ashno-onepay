package tx

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// TxManager aims at facilitating business transactions while abstracting the underlying mechanism,
// be it a database transaction or another transaction mechanism. This allows services to execute
// multiple business use-cases and easily rollback changes in case of error, without creating a
// dependency to the database layer.
//
// Sessions should be constituted of a root tx created with a "New"-type constructor and allow
// the creation of child sessions with `Begin()` and `Transaction()`. Nested transactions should be supported
// as well.
type TxManager interface {
	// Begin returns a new tx with the given context and a started transaction.
	// Using the returned tx should have no side-effect on the parent tx.
	// The underlying transaction mechanism is injected as a value into the new tx's context.
	Begin(ctx context.Context) (TxManager, error)

	// Transaction executes a transaction. If the given function returns an error, the transaction
	// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
	// The underlying transaction mechanism is injected into the context as a value.
	Transaction(ctx context.Context, f func(context.Context) error) error

	// Rollback the changes in the transaction. This action is final.
	Rollback() error

	// Commit the changes in the transaction. This action is final.
	Commit() error

	// Context returns the tx's context. If it's the root tx, `context.Background()` is returned.
	// If it's a child tx started with `Begin()`, then the context will contain the associated
	// transaction mechanism as a value.
	Context() context.Context
}

// Gorm tx implementation.
type Gorm struct {
	db        *gorm.DB
	TxOptions *sql.TxOptions
	ctx       context.Context
}

// GORM create a new root tx for Gorm.
// The transaction options are optional.
func GORM(db *gorm.DB, opt *sql.TxOptions) Gorm {
	return Gorm{
		db:        db,
		TxOptions: opt,
		ctx:       context.Background(),
	}
}

// Begin returns a new tx with the given context and a started DBGorm transaction.
// The returned tx has manual controls. Make sure a call to `Rollback()` or `Commit()`
// is executed before the tx is expired (eligible for garbage collection).
// The Gorm DBGorm associated with this tx is injected as a value into the new tx's context.
// If a Gorm DBGorm is found in the given context, it will be used instead of this TxManager's DBGorm, allowing for
// nested transactions.
func (s Gorm) Begin(ctx context.Context) (TxManager, error) {
	tx := DBGorm(ctx, s.db).WithContext(ctx).Begin(s.TxOptions)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return Gorm{
		ctx:       context.WithValue(ctx, dbKey{}, tx),
		TxOptions: s.TxOptions,
		db:        tx,
	}, nil
}

// Rollback the changes in the transaction. This action is final.
func (s Gorm) Rollback() error {
	return s.db.Rollback().Error
}

// Commit the changes in the transaction. This action is final.
func (s Gorm) Commit() error {
	return s.db.Commit().Error
}

// Context returns the tx's context. If it's the root tx, `context.Background()`
// is returned. If it's a child tx started with `Begin()`, then the context will contain
// the associated Gorm DBGorm and can be used in combination with `tx.DBGorm()`.
func (s Gorm) Context() context.Context {
	return s.ctx
}

// dbKey the key used to store the database in the context.
type dbKey struct{}

// Transaction executes a transaction. If the given function returns an error, the transaction
// is rolled back. Otherwise it is automatically committed before `Transaction()` returns.
//
// The Gorm DBGorm associated with this tx is injected into the context as a value so `tx.DBGorm()`
// can be used to retrieve it.
func (s Gorm) Transaction(ctx context.Context, f func(context.Context) error) error {
	tx := DBGorm(ctx, s.db).WithContext(ctx).Begin(s.TxOptions)
	if tx.Error != nil {
		return tx.Error
	}
	c := context.WithValue(ctx, dbKey{}, tx)
	err := f(c)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// DBGorm returns the Gorm instance stored in the given context. Returns the given fallback
// if no Gorm DBGorm could be found in the context.
func DBGorm(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	db := ctx.Value(dbKey{})
	if db == nil {
		return fallback
	}
	return db.(*gorm.DB)
}
