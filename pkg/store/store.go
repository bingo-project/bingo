package store

import (
	"context"
	"errors"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bingo-project/bingo/pkg/store/logger/empty"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// DBProvider defines an interface for providing a database connection.
type DBProvider interface {
	// DB returns the database instance for the given context.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
}

// Option defines a function type for configuring the Store.
type Option[T any] func(*Store[T])

// Store represents a generic data store with logging capabilities.
type Store[T any] struct {
	logger  Logger
	storage DBProvider
}

// WithLogger returns an Option function that sets the provided Logger to the Store for logging purposes.
func WithLogger[T any](logger Logger) Option[T] {
	return func(s *Store[T]) {
		s.logger = logger
	}
}

// NewStore creates a new instance of Store with the provided DBProvider.
func NewStore[T any](storage DBProvider, logger Logger) *Store[T] {
	if logger == nil {
		logger = empty.NewLogger()
	}

	return &Store[T]{
		logger:  logger,
		storage: storage,
	}
}

// db retrieves the database instance and applies the provided where conditions.
func (s *Store[T]) db(ctx context.Context, wheres ...where.Where) *gorm.DB {
	dbInstance := s.storage.DB(ctx)
	for _, whr := range wheres {
		if whr != nil {
			dbInstance = whr.Where(dbInstance)
		}
	}

	return dbInstance
}

// DB returns the database instance for external use.
func (s *Store[T]) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	return s.db(ctx, wheres...)
}

// Create inserts a new object into the database.
func (s *Store[T]) Create(ctx context.Context, obj *T) error {
	if err := s.db(ctx).Create(obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to insert object into database", "object", obj)

		return err
	}

	return nil
}

// Update modifies an existing object in the database.
func (s *Store[T]) Update(ctx context.Context, obj *T, fields ...string) error {
	if err := s.db(ctx).Select(fields).Save(obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to update object in database", "object", obj)

		return err
	}

	return nil
}

// Delete removes an object from the database based on the provided where options.
func (s *Store[T]) Delete(ctx context.Context, opts *where.Options) error {
	err := s.db(ctx, opts).Delete(new(T)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error(ctx, err, "Failed to delete object from database", "conditions", opts)

		return err
	}

	return nil
}

// Get retrieves a single object from the database based on the provided where options.
func (s *Store[T]) Get(ctx context.Context, opts *where.Options) (*T, error) {
	var obj T
	if err := s.db(ctx, opts).First(&obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to retrieve object from database", "conditions", opts)

		return nil, err
	}

	return &obj, nil
}

// List retrieves a list of objects from the database based on the provided where options.
func (s *Store[T]) List(ctx context.Context, opts *where.Options) (count int64, ret []*T, err error) {
	err = s.db(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	if err != nil {
		s.logger.Error(ctx, err, "Failed to list objects from database", "conditions", opts)
	}

	return
}

func (s *Store[T]) Find(ctx context.Context, opts *where.Options) (ret []*T, err error) {
	err = s.db(ctx, opts).Order("id desc").Find(&ret).Error
	if err != nil {
		s.logger.Error(ctx, err, "Failed to list objects from database", "conditions", opts)
	}

	return
}

// Last retrieves a single object from the database based on the provided where options.
func (s *Store[T]) Last(ctx context.Context, opts *where.Options) (*T, error) {
	var obj T
	if err := s.db(ctx, opts).Last(&obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to retrieve object from database", "conditions", opts)

		return nil, err
	}

	return &obj, nil
}

// CreateInBatch inserts objects into the database in batch.
func (s *Store[T]) CreateInBatch(ctx context.Context, objs []*T, batchSize int) error {
	if err := s.db(ctx).CreateInBatches(objs, batchSize).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to insert objects into database in batch", "count", len(objs))

		return err
	}

	return nil
}

// Upsert creates or modifies an existing object in the database.
func (s *Store[T]) Upsert(ctx context.Context, obj *T, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	if err := s.db(ctx).Clauses(do).Create(obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to upsert object in database", "object", obj)

		return err
	}

	return nil
}

// CreateIfNotExist creates the object if it does not exist.
func (s *Store[T]) CreateIfNotExist(ctx context.Context, obj *T) error {
	db := s.db(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(obj)
	if err := db.Error; err != nil {
		s.logger.Error(ctx, err, "Failed to insert object into database if not exists", "object", obj)

		return err
	}

	return nil
}

// FirstOrCreate finds or creates the object based on the condition.
func (s *Store[T]) FirstOrCreate(ctx context.Context, where any, obj *T) error {
	db := s.db(ctx).Where(where).Attrs(obj).FirstOrCreate(obj)
	if err := db.Error; err != nil {
		s.logger.Error(ctx, err, "Failed to find or create object in database", "condition", where, "object", obj)

		return err
	}

	return nil
}

// UpdateOrCreate updates or creates the object in a transaction with locking.
func (s *Store[T]) UpdateOrCreate(ctx context.Context, where any, obj *T) error {
	err := s.db(ctx).Transaction(func(tx *gorm.DB) error {
		return s.updateOrCreateInTx(tx, where, obj)
	})
	if err != nil {
		s.logger.Error(ctx, err, "Failed to update or create object in database", "condition", where, "object", obj)
	}

	return err
}

// updateOrCreateInTx performs the actual update or create logic in a transaction.
func (s *Store[T]) updateOrCreateInTx(tx *gorm.DB, where any, obj *T) error {
	var exist T
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(where).
		First(&exist).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// If record exists, set the ID from the existing record using reflection
	if err == nil {
		s.setIDFromExisting(obj, exist)
	}

	return tx.Omit("CreatedAt").Save(obj).Error
}

// setIDFromExisting copies the ID field from the existing record to the new object using reflection.
func (s *Store[T]) setIDFromExisting(obj *T, exist T) {
	objValue := reflect.ValueOf(obj).Elem()
	existValue := reflect.ValueOf(exist)
	idField := existValue.FieldByName("ID")
	if !idField.IsValid() {
		return
	}

	idFieldPtr := objValue.FieldByName("ID")
	if idFieldPtr.CanSet() {
		idFieldPtr.Set(idField)
	}
}

// DeleteInBatch deletes objects by IDs.
func (s *Store[T]) DeleteInBatch(ctx context.Context, ids []uint) error {
	db := s.db(ctx).Where("id IN (?)", ids).Delete(new(T))
	if err := db.Error; err != nil {
		s.logger.Error(ctx, err, "Failed to delete objects in batch", "count", len(ids))

		return err
	}

	return nil
}
