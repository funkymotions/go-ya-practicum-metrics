package service

import "strconv"

type metricRepoInterface interface {
	SetGauge(name string, parameter float64)
	SetCounter(name string, parameter int64)
}

type metricService struct {
	repo metricRepoInterface
}

func NewMetricService(repo metricRepoInterface) *metricService {
	return &metricService{
		repo: repo,
	}
}

func (s *metricService) SetCounter(name string, rawValue string) error {
	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return err
	}
	s.repo.SetCounter(name, value)
	return nil
}

func (s *metricService) SetGauge(name string, rawValue string) error {
	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return err
	}
	s.repo.SetGauge(name, value)
	return nil
}
