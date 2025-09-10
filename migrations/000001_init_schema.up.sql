-- creates table with metric types
CREATE TABLE IF NOT EXISTS metric_types (
  id SERIAL PRIMARY KEY,
  metric_type VARCHAR(255) UNIQUE NOT NULL
);

-- creates table with metrics data
CREATE TABLE IF NOT EXISTS metrics (
  id VARCHAR(255) NOT NULL,
  metric_type_id INT REFERENCES metric_types(id),
  value DOUBLE PRECISION,
  delta BIGINT,
  hash VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT metrics_pkey PRIMARY KEY (id, metric_type_id)
);

-- fixtures
INSERT INTO metric_types (metric_type) VALUES
  ('gauge'),
  ('counter')
ON CONFLICT (metric_type) DO NOTHING;
