package repository

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/driver"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go.uber.org/zap"
)

type metricRepository struct {
	memStorage    map[string]models.Metrics
	mu            sync.RWMutex
	writeInterval time.Duration
	filePath      string
	stopCh        chan struct{}
	doneCh        chan struct{}
	driver        *driver.SQLDriver
	logger        *zap.Logger
	gaugeTypeID   uint
	counterTypeID uint
	metricTypes   map[uint]string
}

func NewMetricRepository(
	filePath string,
	isRestoreNeeded bool,
	writeInterval time.Duration,
	db *driver.SQLDriver,
	stopCh chan struct{},
	doneCh chan struct{},

) *metricRepository {
	l, _ := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	r := &metricRepository{
		memStorage:    make(map[string]models.Metrics),
		mu:            sync.RWMutex{},
		writeInterval: writeInterval,
		filePath:      filePath,
		driver:        db,
		stopCh:        stopCh,
		doneCh:        doneCh,
		logger:        l,
	}
	if isRestoreNeeded {
		r.readMetricsFromFile()
	}
	if writeInterval > 0 && filePath != "" {
		ticker := time.NewTicker(writeInterval)
		go func() {
			defer close(r.doneCh)
			for {
				select {
				case <-ticker.C:
					r.writeMetricsToFile()
				case <-r.stopCh:
					ticker.Stop()
					return
				}
			}
		}()
	}
	if db != nil {
		r.initDBSchema()
	}
	return r
}

func (r *metricRepository) SetGaugeIntrospect(name string, value float64) {
	if r.driver == nil {
		r.logger.Warn("DB is not initialized. Continuing...")
		r.SetGauge(name, value)
	} else {
		err := r.upsertMetric(
			nil,
			&models.Metrics{
				ID:    name,
				MType: models.Gauge,
				Value: &value,
				Hash:  "",
			})
		if err != nil {
			r.logger.Error("Error upserting gauge metric to DB:", zap.Error(err))
		}
		// fallback to in-memory storage
		r.SetGauge(name, value)
	}
}

func (r *metricRepository) SetCounterIntrospect(name string, delta int64) {
	if r.driver == nil {
		r.logger.Warn("DB is not initialized. Continuing...")
		r.SetCounter(name, delta)
	} else {
		err := r.upsertMetric(
			nil,
			&models.Metrics{
				ID:    name,
				MType: models.Counter,
				Delta: &delta,
				Hash:  "",
			})
		if err != nil {
			r.logger.Error("Error upserting counter metric to DB:", zap.Error(err))
		}
		// fallback to in-memory storage
		r.SetCounter(name, delta)
	}
}

func (r *metricRepository) SetGauge(name string, value float64) {
	r.mu.Lock()
	key := models.Gauge + ":" + name
	m, exists := r.memStorage[key]
	if !exists {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
			Hash:  "",
		}
		r.memStorage[key] = metric
	} else {
		m.Value = &value
		r.memStorage[key] = m
	}
	r.mu.Unlock()
	// write metrics to disk in same request goroutine
	if r.writeInterval == 0 {
		r.writeMetricsToFile()
	}
}

func (r *metricRepository) SetCounter(name string, delta int64) {
	r.mu.Lock()
	key := models.Counter + ":" + name
	m, exists := r.memStorage[key]
	if !exists {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Value: nil,
			Delta: &delta,
			Hash:  "",
		}
		r.memStorage[key] = metric
	} else {
		*m.Delta += delta
		r.memStorage[key] = m
	}
	r.mu.Unlock()
	// write metrics to disk in same request goroutine
	if r.writeInterval == 0 {
		r.writeMetricsToFile()
	}
}

func (r *metricRepository) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := metricType + ":" + name
	m, exists := r.memStorage[key]
	if !exists && r.driver != nil {
		var typeID uint
		switch metricType {
		case models.Gauge:
			typeID = r.gaugeTypeID
		case models.Counter:
			typeID = r.counterTypeID
		default:
			return nil, false
		}
		metric, err := r.readMetricFromDB(&models.Metrics{
			ID:      name,
			MTypeID: typeID,
		})
		if err != nil {
			return nil, false
		}
		m = *metric
		exists = true
	}
	return &m, exists
}

func (r *metricRepository) GetAllMetrics() map[string]models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.memStorage
}

func (r *metricRepository) SetMetricBulk(m *[]models.Metrics) error {
	if r.driver == nil {
		for _, metric := range *m {
			switch metric.MType {
			case models.Gauge:
				r.SetGauge(metric.ID, *metric.Value)
			case models.Counter:
				r.SetCounter(metric.ID, *metric.Delta)
			}
		}
	} else {
		tx, err := r.driver.DB.Begin()
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					r.logger.Error("tx rollback error:", zap.Error(rbErr))
				}
				return
			}
			if err := tx.Commit(); err != nil {
				r.logger.Error("Error committing transaction:", zap.Error(err))
			}
		}()
		if err != nil {
			r.logger.Error("Error beginning transaction:", zap.Error(err))
			return err
		}
		for _, metric := range *m {
			err := r.upsertMetric(tx, &metric)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// donno where to place this method for now
// repo will be used to host db connection wrapper,
// so in case of responsility separation it should be rignt place
// Panic recovery added to avoid server crash in case of empty DSN server init to support older tests
func (r *metricRepository) Ping() (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("Recovered in Ping", zap.Any("panic", rec))
			if e, ok := rec.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("unknown error: %v", rec)
			}
		}
	}()
	return r.driver.DB.Ping()
}

func (r *metricRepository) readMetricFromDB(m *models.Metrics) (*models.Metrics, error) {
	if r.driver == nil {
		return nil, fmt.Errorf("DB is not initialized")
	}
	sql := "SELECT id, metric_type_id, delta, value FROM metrics WHERE id=$1 AND metric_type_id=$2 LIMIT 1;"
	row := r.driver.DB.QueryRow(sql, m.ID, m.MTypeID)
	var result models.Metrics
	if err := row.Scan(&result.ID, &result.MTypeID, &result.Delta, &result.Value); err != nil {
		return nil, err
	}
	result.MType = r.metricTypes[result.MTypeID]
	return &result, nil
}

func (r *metricRepository) upsertMetric(tx *sql.Tx, m *models.Metrics) error {
	if r.driver == nil {
		return fmt.Errorf("DB is not initialized")
	}
	r.logger.Info("Upserting metric to DB", zap.String("metric", m.String()))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var query string
	var typeID uint
	var value interface{}
	switch m.MType {
	case models.Counter:
		typeID = r.counterTypeID
		value = m.Delta
		query = `
			INSERT INTO
			metrics
				(id, metric_type_id, delta)
			VALUES
				($1, $2, $3)
			ON CONFLICT (id, metric_type_id)
			DO UPDATE SET
				delta = metrics.delta + EXCLUDED.delta,
				updated_at = NOW();
		`
	case models.Gauge:
		value = m.Value
		typeID = r.gaugeTypeID
		query = `
			INSERT INTO
			metrics
				(id, metric_type_id, value)
			VALUES
				($1, $2, $3)
			ON CONFLICT (id, metric_type_id)
			DO UPDATE SET
				value = EXCLUDED.value,
				updated_at = NOW();
		`
	}
	if tx != nil {
		_, err := tx.Exec(query, m.ID, typeID, value)
		return err
	} else {
		_, err := r.driver.DB.ExecContext(ctx, query, m.ID, typeID, value)
		return err
	}
}

func (r *metricRepository) writeMetricsToFile() error {
	if len(r.memStorage) == 0 {
		return nil
	}
	f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	defer buf.Flush()
	r.mu.RLock()
	defer r.mu.RUnlock()
	var metrics []models.Metrics
	for _, m := range r.memStorage {
		metrics = append(metrics, m)
	}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metrics); err != nil {
		return err
	}
	return nil
}

func (r *metricRepository) readMetricsFromFile() error {
	f, err := os.OpenFile(r.filePath, os.O_RDONLY, 0444)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	var metrics []models.Metrics
	if err := json.NewDecoder(buf).Decode(&metrics); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range metrics {
		key := m.MType + ":" + m.ID
		r.memStorage[key] = m
	}
	return nil
}

func (r *metricRepository) initDBSchema() {
	if r.driver == nil {
		r.logger.Warn("DB can not be initialized")
		return
	}
	driver, err := postgres.WithInstance(r.driver.DB, &postgres.Config{})
	if err != nil {
		r.logger.Error("Error creating migrate postgres instance:", zap.Error(err))
		return
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		r.logger.Error("Error creating golang-migrate instance:", zap.Error(err))
		return
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		r.logger.Error("Error applying migrations:", zap.Error(err))
		return
	}
	r.cacheMetricTypeIDs()
	r.logger.Info("Database migrations applied successfully")
}

func (r *metricRepository) cacheMetricTypeIDs() error {
	if r.driver == nil {
		err := fmt.Errorf("DB is not initialized")
		r.logger.Error("Error caching metric type IDs", zap.Error(err))
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	query := `SELECT id, metric_type FROM metric_types;`
	rows, err := r.driver.DB.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Error querying metric types:", zap.Error(err))
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uint
		var typeName string
		if err := rows.Scan(&id, &typeName); err != nil {
			return err
		}
		switch typeName {
		case models.Gauge:
			r.gaugeTypeID = id
		case models.Counter:
			r.counterTypeID = id
		}
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Error scanning metric types:", zap.Error(err))
		return err
	}
	r.logger.Info("Cached metric type IDs",
		zap.Uint("gauge_type_id", r.gaugeTypeID),
		zap.Uint("counter_type_id", r.counterTypeID),
	)
	r.metricTypes = map[uint]string{
		r.gaugeTypeID:   models.Gauge,
		r.counterTypeID: models.Counter,
	}
	return nil
}
